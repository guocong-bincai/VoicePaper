package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserPointsRepository struct {
	db *gorm.DB
}

func NewUserPointsRepository(db *gorm.DB) *UserPointsRepository {
	return &UserPointsRepository{db: db}
}

// GetByUserID 根据用户ID获取积分信息
func (r *UserPointsRepository) GetByUserID(userID uint) (*model.UserPoints, error) {
	var userPoints model.UserPoints
	err := r.db.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		return nil, err
	}
	return &userPoints, nil
}

// Create 创建用户积分记录
func (r *UserPointsRepository) Create(userPoints *model.UserPoints) error {
	return r.db.Create(userPoints).Error
}

// Update 更新用户积分信息
func (r *UserPointsRepository) Update(userPoints *model.UserPoints) error {
	return r.db.Save(userPoints).Error
}

// AddPoints 增加积分（事务中执行）
func (r *UserPointsRepository) AddPoints(tx *gorm.DB, userID uint, points int) (*model.UserPoints, error) {
	var userPoints model.UserPoints

	// 使用行锁防止并发问题
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userPoints).Error

	if err != nil {
		return nil, err
	}

	// 更新积分
	userPoints.TotalPoints += points
	userPoints.CurrentPoints += points

	// 重新计算等级
	userPoints.CalculateLevel()

	// 保存
	err = tx.Save(&userPoints).Error
	if err != nil {
		return nil, err
	}

	return &userPoints, nil
}

// IncrementArticlesRead 增加阅读文章计数
func (r *UserPointsRepository) IncrementArticlesRead(userID uint) error {
	return r.db.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		UpdateColumn("total_articles_read", gorm.Expr("total_articles_read + ?", 1)).
		Error
}

// IncrementDictationsCompleted 增加默写完成计数
func (r *UserPointsRepository) IncrementDictationsCompleted(userID uint) error {
	return r.db.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		UpdateColumn("total_dictations_completed", gorm.Expr("total_dictations_completed + ?", 1)).
		Error
}

// UpdateCheckInStats 更新签到统计（在事务中执行）
func (r *UserPointsRepository) UpdateCheckInStats(tx *gorm.DB, userID uint, continuousDays int) error {
	// 一次性更新所有字段，使用 CASE WHEN 来更新最大连续签到天数
	return tx.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"total_check_ins":          gorm.Expr("total_check_ins + ?", 1),
			"continuous_check_ins":     continuousDays,
			"max_continuous_check_ins": gorm.Expr("CASE WHEN ? > max_continuous_check_ins THEN ? ELSE max_continuous_check_ins END", continuousDays, continuousDays),
		}).Error
}

// GetTopUsers 获取积分排行榜
func (r *UserPointsRepository) GetTopUsers(limit int) ([]model.UserPoints, error) {
	var users []model.UserPoints
	err := r.db.Order("total_points DESC").
		Limit(limit).
		Preload("User").
		Find(&users).Error
	return users, err
}

// GetByUserIDWithTx 在事务中根据用户ID获取积分信息
func (r *UserPointsRepository) GetByUserIDWithTx(tx *gorm.DB, userID uint) (*model.UserPoints, error) {
	var userPoints model.UserPoints
	err := tx.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		return nil, err
	}
	return &userPoints, nil
}

// DeductPoints 扣除积分（事务中执行）
func (r *UserPointsRepository) DeductPoints(tx *gorm.DB, userID uint, points int) (*model.UserPoints, error) {
	var userPoints model.UserPoints

	// 使用行锁防止并发问题
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		First(&userPoints).Error

	if err != nil {
		return nil, err
	}

	if userPoints.CurrentPoints < points {
		return nil, gorm.ErrInvalidValue
	}

	// 扣除积分（只扣当前积分，不减累计积分）
	userPoints.CurrentPoints -= points

	// 保存
	err = tx.Save(&userPoints).Error
	if err != nil {
		return nil, err
	}

	return &userPoints, nil
}

// AddMakeupCards 增加补签卡数量（事务中执行）
func (r *UserPointsRepository) AddMakeupCards(tx *gorm.DB, userID uint, count int) error {
	return tx.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		Update("makeup_cards", gorm.Expr("makeup_cards + ?", count)).Error
}

// DeductMakeupCards 扣除补签卡数量（事务中执行）
func (r *UserPointsRepository) DeductMakeupCards(tx *gorm.DB, userID uint, count int) error {
	// 先检查是否足够
	var userPoints model.UserPoints
	err := tx.Where("user_id = ?", userID).First(&userPoints).Error
	if err != nil {
		return err
	}
	if userPoints.MakeupCards < count {
		return gorm.ErrInvalidValue
	}

	return tx.Model(&model.UserPoints{}).
		Where("user_id = ?", userID).
		Update("makeup_cards", gorm.Expr("makeup_cards - ?", count)).Error
}
