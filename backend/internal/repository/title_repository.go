package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

type TitleRepository struct {
	db *gorm.DB
}

func NewTitleRepository(db *gorm.DB) *TitleRepository {
	return &TitleRepository{db: db}
}

// ============ TitleConfig 相关 ============

// GetAllTitleConfigs 获取所有称号配置
func (r *TitleRepository) GetAllTitleConfigs() ([]model.TitleConfig, error) {
	var configs []model.TitleConfig
	err := r.db.Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&configs).Error
	return configs, err
}

// GetTitleConfigByKey 根据key获取称号配置
func (r *TitleRepository) GetTitleConfigByKey(key string) (*model.TitleConfig, error) {
	var config model.TitleConfig
	err := r.db.Where("title_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetTitleConfigsByCategory 根据分类获取称号配置
func (r *TitleRepository) GetTitleConfigsByCategory(category model.TitleCategory) ([]model.TitleConfig, error) {
	var configs []model.TitleConfig
	err := r.db.Where("category = ? AND is_active = ?", category, true).
		Order("sort_order ASC").
		Find(&configs).Error
	return configs, err
}

// ============ UserTitle 相关 ============

// CreateUserTitle 创建用户称号（在事务中执行）
func (r *TitleRepository) CreateUserTitle(tx *gorm.DB, userTitle *model.UserTitle) error {
	return tx.Create(userTitle).Error
}

// GetUserTitles 获取用户已获得的所有称号
func (r *TitleRepository) GetUserTitles(userID uint) ([]model.UserTitle, error) {
	var userTitles []model.UserTitle
	err := r.db.Where("user_id = ?", userID).
		Preload("TitleConfig").
		Order("awarded_at DESC").
		Find(&userTitles).Error
	return userTitles, err
}

// CheckUserHasTitle 检查用户是否已拥有该称号
func (r *TitleRepository) CheckUserHasTitle(userID, titleConfigID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserTitle{}).
		Where("user_id = ? AND title_config_id = ?", userID, titleConfigID).
		Count(&count).Error
	return count > 0, err
}

// GetEquippedTitle 获取用户当前佩戴的称号
func (r *TitleRepository) GetEquippedTitle(userID uint) (*model.UserTitle, error) {
	var userTitle model.UserTitle
	err := r.db.Where("user_id = ? AND is_equipped = ?", userID, true).
		Preload("TitleConfig").
		First(&userTitle).Error
	if err != nil {
		return nil, err
	}
	return &userTitle, nil
}

// EquipTitle 佩戴称号
func (r *TitleRepository) EquipTitle(userID, titleConfigID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 取消所有已佩戴的称号
		err := tx.Model(&model.UserTitle{}).
			Where("user_id = ? AND is_equipped = ?", userID, true).
			Update("is_equipped", false).Error
		if err != nil {
			return err
		}

		// 2. 佩戴新称号
		err = tx.Model(&model.UserTitle{}).
			Where("user_id = ? AND title_config_id = ?", userID, titleConfigID).
			Update("is_equipped", true).Error
		return err
	})
}

// UnequipTitle 取消佩戴称号
func (r *TitleRepository) UnequipTitle(userID uint) error {
	return r.db.Model(&model.UserTitle{}).
		Where("user_id = ? AND is_equipped = ?", userID, true).
		Update("is_equipped", false).Error
}

// GetTitleProgress 获取用户的称号进度（还差多少可以获得某个称号）
func (r *TitleRepository) GetTitleProgress(userID uint) ([]map[string]interface{}, error) {
	// 这个方法需要结合UserPoints数据来计算
	// 在Service层实现更合适
	return nil, nil
}
