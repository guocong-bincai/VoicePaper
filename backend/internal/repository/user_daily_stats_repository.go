package repository

import (
	"time"
	"voicepaper/internal/model"

	"gorm.io/gorm"
)

type UserDailyStatsRepository struct {
	db *gorm.DB
}

func NewUserDailyStatsRepository(db *gorm.DB) *UserDailyStatsRepository {
	return &UserDailyStatsRepository{db: db}
}

// GetOrCreateTodayStats 获取或创建今日统计
func (r *UserDailyStatsRepository) GetOrCreateTodayStats(userID uint) (*model.UserDailyStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var stats model.UserDailyStats
	err := r.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新记录
		stats = model.UserDailyStats{
			UserID:   userID,
			StatDate: today,
		}
		err = r.db.Create(&stats).Error
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &stats, nil
}

// UpdateDuration 更新今日时长（累加）
func (r *UserDailyStatsRepository) UpdateDuration(userID uint, durationSeconds int) error {
	stats, err := r.GetOrCreateTodayStats(userID)
	if err != nil {
		return err
	}

	// 使用 GORM 的 UpdateColumn 来累加时长，而不是覆盖
	return r.db.Model(&model.UserDailyStats{}).
		Where("id = ?", stats.ID).
		UpdateColumn("total_duration_seconds", gorm.Expr("total_duration_seconds + ?", durationSeconds)).
		Error
}

// IncrementNewWords 增加新学单词数
func (r *UserDailyStatsRepository) IncrementNewWords(userID uint, count int) error {
	stats, err := r.GetOrCreateTodayStats(userID)
	if err != nil {
		return err
	}

	// 使用 Updates 和 map 确保只更新指定字段，不影响其他字段
	return r.db.Model(&model.UserDailyStats{}).
		Where("id = ?", stats.ID).
		Updates(map[string]interface{}{
			"new_words": gorm.Expr("COALESCE(new_words, 0) + ?", count),
		}).Error
}

// IncrementReviewedWords 增加复习单词数
func (r *UserDailyStatsRepository) IncrementReviewedWords(userID uint, count int) error {
	stats, err := r.GetOrCreateTodayStats(userID)
	if err != nil {
		return err
	}

	// 使用 Updates 和 map 确保只更新指定字段
	return r.db.Model(&model.UserDailyStats{}).
		Where("id = ?", stats.ID).
		Updates(map[string]interface{}{
			"reviewed_words": gorm.Expr("COALESCE(reviewed_words, 0) + ?", count),
		}).Error
}

// RecordAttempt 记录一次尝试（正确或错误）
func (r *UserDailyStatsRepository) RecordAttempt(userID uint, isCorrect bool) error {
	stats, err := r.GetOrCreateTodayStats(userID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"total_attempts": gorm.Expr("COALESCE(total_attempts, 0) + 1"),
	}

	if isCorrect {
		updates["correct_count"] = gorm.Expr("COALESCE(correct_count, 0) + 1")
	}

	return r.db.Model(&model.UserDailyStats{}).
		Where("id = ?", stats.ID).
		Updates(updates).Error
}

// GetTodayStats 获取今日统计
func (r *UserDailyStatsRepository) GetTodayStats(userID uint) (*model.UserDailyStats, error) {
	return r.GetOrCreateTodayStats(userID)
}

// GetStatsByDate 获取指定日期的统计
func (r *UserDailyStatsRepository) GetStatsByDate(userID uint, date time.Time) (*model.UserDailyStats, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var stats model.UserDailyStats
	err := r.db.Where("user_id = ? AND stat_date = ?", userID, startOfDay).First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// 返回空统计
		return &model.UserDailyStats{
			UserID:   userID,
			StatDate: startOfDay,
		}, nil
	}

	return &stats, err
}

// GetStatsRange 获取日期范围内的统计
func (r *UserDailyStatsRepository) GetStatsRange(userID uint, startDate, endDate time.Time) ([]model.UserDailyStats, error) {
	var stats []model.UserDailyStats
	err := r.db.Where("user_id = ? AND stat_date >= ? AND stat_date <= ?", userID, startDate, endDate).
		Order("stat_date DESC").
		Find(&stats).Error
	return stats, err
}
