package repository

import (
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

// DictationProgressRepository 默写进度仓库
type DictationProgressRepository struct {
	db *gorm.DB
}

// NewDictationProgressRepository 创建默写进度仓库
func NewDictationProgressRepository(db *gorm.DB) *DictationProgressRepository {
	return &DictationProgressRepository{db: db}
}

// GetProgress 获取用户的默写进度
func (r *DictationProgressRepository) GetProgress(userID, articleID uint, dictationType string) (*model.UserDictationProgress, error) {
	var progress model.UserDictationProgress
	err := r.db.Where("user_id = ? AND article_id = ? AND dictation_type = ?", userID, articleID, dictationType).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// SaveProgress 保存或更新默写进度
func (r *DictationProgressRepository) SaveProgress(progress *model.UserDictationProgress) error {
	progress.LastPracticeAt = time.Now()
	progress.UpdatedAt = time.Now()

	// 使用 OnConflict 来处理唯一索引冲突
	return r.db.Save(progress).Error
}

// CreateOrUpdateProgress 创建或更新进度
func (r *DictationProgressRepository) CreateOrUpdateProgress(userID, articleID uint, dictationType string, currentIndex, totalItems, score int, completed bool) error {
	progress := &model.UserDictationProgress{
		UserID:         userID,
		ArticleID:      articleID,
		DictationType:  dictationType,
		CurrentIndex:   currentIndex,
		TotalItems:     totalItems,
		Score:          score,
		Completed:      completed,
		LastPracticeAt: time.Now(),
	}

	// 先尝试查找现有记录
	var existing model.UserDictationProgress
	err := r.db.Where("user_id = ? AND article_id = ? AND dictation_type = ?", userID, articleID, dictationType).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		progress.CreatedAt = time.Now()
		progress.UpdatedAt = time.Now()
		return r.db.Create(progress).Error
	} else if err != nil {
		return err
	}

	// 存在，更新记录
	existing.CurrentIndex = currentIndex
	existing.TotalItems = totalItems
	existing.Score = score
	existing.Completed = completed
	existing.LastPracticeAt = time.Now()
	existing.UpdatedAt = time.Now()
	return r.db.Save(&existing).Error
}

// ResetProgress 重置进度（重新开始）
func (r *DictationProgressRepository) ResetProgress(userID, articleID uint, dictationType string) error {
	return r.db.Where("user_id = ? AND article_id = ? AND dictation_type = ?", userID, articleID, dictationType).
		Delete(&model.UserDictationProgress{}).Error
}
