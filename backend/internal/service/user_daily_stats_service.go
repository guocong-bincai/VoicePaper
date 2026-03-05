package service

import (
	"time"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"gorm.io/gorm"
)

type UserDailyStatsService struct {
	db        *gorm.DB
	statsRepo *repository.UserDailyStatsRepository
	vocabRepo *repository.VocabularyRepository
}

func NewUserDailyStatsService(db *gorm.DB) *UserDailyStatsService {
	return &UserDailyStatsService{
		db:        db,
		statsRepo: repository.NewUserDailyStatsRepository(db),
		vocabRepo: repository.NewVocabularyRepository(),
	}
}

// SyncDuration 同步今日时长
func (s *UserDailyStatsService) SyncDuration(userID uint, durationSeconds int) error {
	return s.statsRepo.UpdateDuration(userID, durationSeconds)
}

// RecordNewWord 记录新学单词
func (s *UserDailyStatsService) RecordNewWord(userID uint) error {
	return s.statsRepo.IncrementNewWords(userID, 1)
}

// RecordReview 记录复习（包含正确率统计）
func (s *UserDailyStatsService) RecordReview(userID uint, isCorrect bool) error {
	// 增加复习数
	err := s.statsRepo.IncrementReviewedWords(userID, 1)
	if err != nil {
		return err
	}

	// 记录正确率
	return s.statsRepo.RecordAttempt(userID, isCorrect)
}

// GetTodayStats 获取今日统计
func (s *UserDailyStatsService) GetTodayStats(userID uint) (*model.UserDailyStats, error) {
	return s.statsRepo.GetTodayStats(userID)
}

// GetTodayStatsWithVocabulary 获取今日统计（合并生词本数据）
func (s *UserDailyStatsService) GetTodayStatsWithVocabulary(userID uint) (map[string]interface{}, error) {
	// 获取基础统计
	stats, err := s.statsRepo.GetTodayStats(userID)
	if err != nil {
		return nil, err
	}

	// 获取生词本今日统计
	today := time.Now()
	vocabStats, err := s.vocabRepo.GetOrCreateDailyStats(userID, today)
	if err != nil {
		vocabStats = &model.VocabularyDailyStats{}
	}

	// 合并数据
	result := map[string]interface{}{
		"duration":         stats.TotalDurationSeconds / 60,                // 转换为分钟
		"duration_seconds": stats.TotalDurationSeconds,                     // 精确秒数
		"new_words":        stats.NewWords + vocabStats.NewWords,           // 合并新学单词
		"reviewed_words":   stats.ReviewedWords + vocabStats.ReviewedWords, // 合并复习单词
		"accuracy":         stats.GetCorrectRate(),                         // 正确率
	}

	return result, nil
}

// GetStatsForDays 获取最近N天的统计
func (s *UserDailyStatsService) GetStatsForDays(userID uint, days int) ([]model.UserDailyStats, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	// 规范化为每天的开始时间
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	return s.statsRepo.GetStatsRange(userID, startDate, endDate)
}
