package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

type PointRecordRepository struct {
	db *gorm.DB
}

func NewPointRecordRepository(db *gorm.DB) *PointRecordRepository {
	return &PointRecordRepository{db: db}
}

// Create 创建积分记录（在事务中执行）
func (r *PointRecordRepository) Create(tx *gorm.DB, record *model.PointRecord) error {
	return tx.Create(record).Error
}

// GetByUserID 获取用户的积分记录列表
func (r *PointRecordRepository) GetByUserID(userID uint, limit, offset int) ([]model.PointRecord, error) {
	var records []model.PointRecord
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error
	return records, err
}

// GetByUserIDAndType 获取用户指定类型的积分记录
func (r *PointRecordRepository) GetByUserIDAndType(userID uint, pointType model.PointType, limit int) ([]model.PointRecord, error) {
	var records []model.PointRecord
	err := r.db.Where("user_id = ? AND type = ?", userID, pointType).
		Order("created_at DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

// GetTotalPointsByType 获取用户某类型的总积分
func (r *PointRecordRepository) GetTotalPointsByType(userID uint, pointType model.PointType) (int, error) {
	var total int64
	err := r.db.Model(&model.PointRecord{}).
		Where("user_id = ? AND type = ?", userID, pointType).
		Select("COALESCE(SUM(points), 0)").
		Scan(&total).Error
	return int(total), err
}

// CountByUserID 获取用户的积分记录总数
func (r *PointRecordRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.PointRecord{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CheckArticleRead 检查用户是否已阅读过该文章（获得过积分）
func (r *PointRecordRepository) CheckArticleRead(userID, articleID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.PointRecord{}).
		Where("user_id = ? AND article_id = ? AND type = ?", userID, articleID, model.PointTypeReadArticle).
		Count(&count).Error
	return count > 0, err
}

// GetRecentRecords 获取最近的积分记录（用于动态展示）
func (r *PointRecordRepository) GetRecentRecords(limit int) ([]model.PointRecord, error) {
	var records []model.PointRecord
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Preload("User").
		Find(&records).Error
	return records, err
}
