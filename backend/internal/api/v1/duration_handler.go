package v1

import (
	"log"
	"net/http"
	"voicepaper/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DurationHandler struct {
	db *gorm.DB
}

func NewDurationHandler(db *gorm.DB) *DurationHandler {
	return &DurationHandler{db: db}
}

// SyncReadingDuration 同步阅读时长
// POST /api/v1/duration/sync
// 将前端的阅读统计时长同步到后端，更新 vp_user_points 表的 total_duration_minutes
func (h *DurationHandler) SyncReadingDuration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		TotalReadingSeconds int64 `json:"total_reading_seconds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	uid := userID.(uint)

	// 转换秒为分钟
	durationMinutes := req.TotalReadingSeconds / 60

	// 使用 Clauses 实现 INSERT ... ON DUPLICATE KEY UPDATE
	userPoints := model.UserPoints{
		UserID:               uid,
		TotalDurationMinutes: durationMinutes,
		Level:                1,
		LevelName:            "初学者",
	}

	// 先尝试更新现有记录
	updateResult := h.db.Where("user_id = ?", uid).Model(&model.UserPoints{}).Update("total_duration_minutes", durationMinutes)

	if updateResult.Error != nil {
		log.Printf("❌ 更新用户时长失败 (user_id=%d): %v", uid, updateResult.Error)
	}

	// 如果没有更新任何记录，则创建新记录
	if updateResult.RowsAffected == 0 {
		if err := h.db.Create(&userPoints).Error; err != nil {
			log.Printf("❌ 创建用户点数记录失败 (user_id=%d): %v", uid, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建记录失败", "details": err.Error()})
			return
		}
	}

	// 也更新 vp_users 表的备份字段
	h.db.Where("id = ?", uid).Model(&model.User{}).Update("total_duration_minutes", durationMinutes)

	log.Printf("✅ 同步阅读时长成功 (user_id=%d, duration=%d分钟)", uid, durationMinutes)

	c.JSON(http.StatusOK, gin.H{
		"message":          "同步成功",
		"duration_minutes": durationMinutes,
		"duration_seconds": req.TotalReadingSeconds,
	})
}

// GetUserDuration 获取用户的累积时长
// GET /api/v1/duration/me
// 返回 vp_user_points 的时长 + vp_user_daily_stats 的累计时长
func (h *DurationHandler) GetUserDuration(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)

	// 1. 获取 vp_user_points 的时长（分钟）
	var userPoints model.UserPoints
	pointsDuration := int64(0)
	if err := h.db.Where("user_id = ?", uid).First(&userPoints).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			return
		}
	} else {
		pointsDuration = userPoints.TotalDurationMinutes
	}

	// 2. 获取 vp_user_daily_stats 的累计时长（秒）
	var result struct {
		TotalSeconds int64
	}

	err := h.db.Table("vp_user_daily_stats").
		Select("COALESCE(SUM(total_duration_seconds), 0) as total_seconds").
		Where("user_id = ?", uid).
		Where("deleted_at IS NULL").
		Scan(&result).Error

	dailyStatsDuration := int64(0)
	if err == nil {
		dailyStatsDuration = result.TotalSeconds / 60 // 转换为分钟
	}

	// 3. 合并两个数据源的时长
	totalDuration := pointsDuration + dailyStatsDuration

	log.Printf("📊 用户 %d 总时长: vp_user_points=%d分钟 + vp_user_daily_stats=%d分钟 = %d分钟",
		uid, pointsDuration, dailyStatsDuration, totalDuration)

	c.JSON(http.StatusOK, gin.H{
		"total_duration_minutes":       totalDuration,      // 总时长（分钟）
		"points_duration_minutes":      pointsDuration,     // user_points 中的时长
		"daily_stats_duration_minutes": dailyStatsDuration, // daily_stats 中的时长
	})
}
