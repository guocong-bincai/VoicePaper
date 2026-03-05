package v1

import (
	"net/http"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DailyStatsHandler struct {
	db      *gorm.DB
	service *service.UserDailyStatsService
}

func NewDailyStatsHandler(db *gorm.DB) *DailyStatsHandler {
	return &DailyStatsHandler{
		db:      db,
		service: service.NewUserDailyStatsService(db),
	}
}

// SyncDuration 同步今日时长
// POST /api/v1/daily-stats/duration
func (h *DailyStatsHandler) SyncDuration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		DurationSeconds int `json:"duration_seconds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	uid := userID.(uint)
	err := h.service.SyncDuration(uid, req.DurationSeconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "同步失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "同步成功",
		"duration_seconds": req.DurationSeconds,
		"duration_minutes": req.DurationSeconds / 60,
	})
}

// RecordNewWord 记录新学单词
// POST /api/v1/daily-stats/new-word
func (h *DailyStatsHandler) RecordNewWord(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	err := h.service.RecordNewWord(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "记录成功"})
}

// RecordReview 记录复习
// POST /api/v1/daily-stats/review
func (h *DailyStatsHandler) RecordReview(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		IsCorrect bool `json:"is_correct"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	uid := userID.(uint)
	err := h.service.RecordReview(uid, req.IsCorrect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "记录成功"})
}

// GetTodayStats 获取今日统计
// GET /api/v1/daily-stats/today
func (h *DailyStatsHandler) GetTodayStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	stats, err := h.service.GetTodayStatsWithVocabulary(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserVocabularyCount 获取用户词汇量
// GET /api/v1/daily-stats/vocabulary-count
// 词汇量 = 生词本单词数 + 单词书已学习的单词数
func (h *DailyStatsHandler) GetUserVocabularyCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)

	// 1. 获取生词本单词数
	var vocabularyCount int64
	h.db.Table("vp_vocabulary").
		Where("user_id = ? AND deleted_at IS NULL", uid).
		Count(&vocabularyCount)

	// 2. 获取单词书已学习的单词数（所有单词书的 current_index 之和）
	var wordbookLearnedCount int64
	h.db.Table("vp_wordbook_progress").
		Select("COALESCE(SUM(current_index), 0)").
		Where("user_id = ? AND deleted_at IS NULL", uid).
		Scan(&wordbookLearnedCount)

	// 3. 总词汇量
	totalVocabulary := vocabularyCount + wordbookLearnedCount

	c.JSON(http.StatusOK, gin.H{
		"total":            totalVocabulary,      // 总词汇量
		"vocabulary_count": vocabularyCount,      // 生词本单词数
		"wordbook_learned": wordbookLearnedCount, // 单词书已学习数
	})
}
