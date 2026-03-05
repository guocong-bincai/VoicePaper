package repository

import (
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReadingProgressRepository struct {
	db *gorm.DB
}

func NewReadingProgressRepository(db *gorm.DB) *ReadingProgressRepository {
	return &ReadingProgressRepository{db: db}
}

// GetByUserAndArticle 获取用户某篇文章的阅读进度
func (r *ReadingProgressRepository) GetByUserAndArticle(userID, articleID uint) (*model.ReadingProgress, error) {
	var progress model.ReadingProgress
	err := r.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// SaveProgress 保存或更新阅读进度
func (r *ReadingProgressRepository) SaveProgress(userID, articleID uint, currentTime, duration, progressPercent float64) (*model.ReadingProgress, error) {
	var progress model.ReadingProgress

	// 尝试获取现有记录
	err := r.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&progress).Error

	isCompleted := progressPercent >= 80.0
	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		progress = model.ReadingProgress{
			UserID:      userID,
			ArticleID:   articleID,
			CurrentTime: currentTime,
			Duration:    duration,
			Progress:    progressPercent,
			IsCompleted: isCompleted,
			ReadCount:   1,
			TotalTime:   int(currentTime),
			LastReadAt:  now,
		}
		err = r.db.Create(&progress).Error
		if err != nil {
			return nil, err
		}
		return &progress, nil
	}

	if err != nil {
		return nil, err
	}

	// 更新现有记录
	updates := map[string]interface{}{
		"current_time": currentTime,
		"duration":     duration,
		"progress":     progressPercent,
		"last_read_at": now,
	}

	// total_time 取当前播放位置和原来的最大值（表示这篇文章最远读到哪里）
	updates["total_time"] = gorm.Expr("GREATEST(total_time, ?)", int(currentTime))

	// 如果首次完成
	if isCompleted && !progress.IsCompleted {
		updates["is_completed"] = true
		updates["read_count"] = gorm.Expr("read_count + ?", 1)
	}

	err = r.db.Model(&progress).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// 重新获取更新后的记录
	err = r.db.Where("user_id = ? AND article_id = ?", userID, articleID).First(&progress).Error
	return &progress, err
}

// GetUserReadingHistory 获取用户阅读历史
func (r *ReadingProgressRepository) GetUserReadingHistory(userID uint, limit, offset int) ([]model.ReadingProgress, int64, error) {
	var progresses []model.ReadingProgress
	var total int64

	// 获取总数
	err := r.db.Model(&model.ReadingProgress{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表，按最后阅读时间倒序
	err = r.db.Where("user_id = ?", userID).
		Order("last_read_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Article").
		Find(&progresses).Error

	return progresses, total, err
}

// GetLastReadArticle 获取用户最后阅读的文章进度
func (r *ReadingProgressRepository) GetLastReadArticle(userID uint) (*model.ReadingProgress, error) {
	var progress model.ReadingProgress
	err := r.db.Where("user_id = ?", userID).
		Order("last_read_at DESC").
		Preload("Article").
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// MarkCompleted 标记文章阅读完成
func (r *ReadingProgressRepository) MarkCompleted(userID, articleID uint) error {
	return r.db.Model(&model.ReadingProgress{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"progress":     100.0,
			"read_count":   gorm.Expr("read_count + ?", 1),
		}).Error
}

// CreateOrUpdate 使用 UPSERT 保存进度（用于高频更新场景）
// 注意：这个方法不更新total_time，需要用SaveProgress来正确累加时间
func (r *ReadingProgressRepository) CreateOrUpdate(progress *model.ReadingProgress) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "article_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"current_time", "duration", "progress", "last_read_at", "updated_at"}),
	}).Create(progress).Error
}

// UpdateTotalTime 更新累计阅读时间（直接设置为当前播放位置）
func (r *ReadingProgressRepository) UpdateTotalTime(userID, articleID uint, currentTime float64) error {
	// total_time 取 current_time 和 原total_time 的最大值
	return r.db.Model(&model.ReadingProgress{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Update("total_time", gorm.Expr("GREATEST(total_time, ?)", int(currentTime))).
		Error
}

// GetUserReadingStats 获取用户阅读统计
func (r *ReadingProgressRepository) GetUserReadingStats(userID uint) (map[string]interface{}, error) {
	var result struct {
		TotalReadingTime int   `json:"total_reading_time"` // 累计阅读时长（秒）
		TotalArticles    int64 `json:"total_articles"`     // 阅读文章数
		CompletedCount   int64 `json:"completed_count"`    // 完成阅读的文章数
	}

	// 获取累计阅读时长和文章数
	err := r.db.Model(&model.ReadingProgress{}).
		Select("COALESCE(SUM(total_time), 0) as total_reading_time, COUNT(*) as total_articles, SUM(CASE WHEN is_completed = 1 THEN 1 ELSE 0 END) as completed_count").
		Where("user_id = ?", userID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_reading_time": result.TotalReadingTime,
		"total_articles":     result.TotalArticles,
		"completed_count":    result.CompletedCount,
	}, nil
}
