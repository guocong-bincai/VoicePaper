package repository

import (
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

// FeedbackRepository 反馈数据仓库
type FeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository 创建反馈仓库实例
func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create 创建反馈
func (r *FeedbackRepository) Create(feedback *model.Feedback) error {
	return r.db.Create(feedback).Error
}

// GetAll 获取所有反馈（管理员功能，暂不实现）
func (r *FeedbackRepository) GetAll(page, pageSize int) ([]model.Feedback, int64, error) {
	var feedbacks []model.Feedback
	var total int64

	// 统计总数
	if err := r.db.Model(&model.Feedback{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&feedbacks).Error; err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}

// GetByID 根据ID获取反馈
func (r *FeedbackRepository) GetByID(id uint) (*model.Feedback, error) {
	var feedback model.Feedback
	if err := r.db.First(&feedback, id).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

// GetByUserID 获取用户的反馈列表
func (r *FeedbackRepository) GetByUserID(userID uint, page, pageSize int) ([]model.Feedback, int64, error) {
	var feedbacks []model.Feedback
	var total int64

	// 统计总数
	if err := r.db.Model(&model.Feedback{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&feedbacks).Error; err != nil {
		return nil, 0, err
	}

	return feedbacks, total, nil
}
