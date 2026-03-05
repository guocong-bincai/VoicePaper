package v1

import (
	"log"
	"net/http"
	"strconv"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PointHandler struct {
	pointService *service.PointService
	db           *gorm.DB
}

func NewPointHandler() *PointHandler {
	return &PointHandler{
		pointService: service.NewPointService(repository.DB),
		db:           repository.DB,
	}
}

// GetMyPoints 获取我的积分信息
// GET /api/v1/points/me
func (h *PointHandler) GetMyPoints(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	userPoints, err := h.pointService.GetUserPoints(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取积分信息失败", "details": err.Error()})
		return
	}

	// 获取佩戴的称号
	equippedTitle, _ := h.pointService.GetEquippedTitle(uid)

	// 获取 vp_user_daily_stats 的累计时长（秒）
	var result struct {
		TotalSeconds int64
	}

	h.db.Table("vp_user_daily_stats").
		Select("COALESCE(SUM(total_duration_seconds), 0) as total_seconds").
		Where("user_id = ?", uid).
		Where("deleted_at IS NULL").
		Scan(&result)

	dailyStatsDuration := result.TotalSeconds / 60 // 转换为分钟
	totalDuration := userPoints.TotalDurationMinutes + dailyStatsDuration

	log.Printf("📊 GetMyPoints - 用户 %d 总时长: vp_user_points=%d + vp_user_daily_stats=%d = %d分钟",
		uid, userPoints.TotalDurationMinutes, dailyStatsDuration, totalDuration)

	// 构建返回数据
	response := gin.H{
		"id":                         userPoints.ID,
		"user_id":                    userPoints.UserID,
		"total_points":               userPoints.TotalPoints,
		"current_points":             userPoints.CurrentPoints,
		"level":                      userPoints.Level,
		"level_name":                 userPoints.LevelName,
		"total_articles_read":        userPoints.TotalArticlesRead,
		"total_dictations_completed": userPoints.TotalDictationsCompleted,
		"total_check_ins":            userPoints.TotalCheckIns,
		"continuous_check_ins":       userPoints.ContinuousCheckIns,
		"max_continuous_check_ins":   userPoints.MaxContinuousCheckIns,
		"total_duration_minutes":     totalDuration, // 合并后的总时长
	}

	// 如果有佩戴的称号，使用称号名称覆盖 level_name
	if equippedTitle != nil {
		response["equipped_title"] = gin.H{
			"id":         equippedTitle.TitleConfig.ID,
			"title_key":  equippedTitle.TitleConfig.TitleKey,
			"title_name": equippedTitle.TitleConfig.TitleName,
			"title_icon": equippedTitle.TitleConfig.TitleIcon,
			"rarity":     equippedTitle.TitleConfig.Rarity,
		}
		// 用佩戴的称号替换 level_name 显示
		response["display_name"] = equippedTitle.TitleConfig.TitleName
		response["display_icon"] = equippedTitle.TitleConfig.TitleIcon
	} else {
		response["display_name"] = userPoints.LevelName
		response["display_icon"] = ""
	}

	c.JSON(http.StatusOK, response)
}

// GetPointRecords 获取积分记录
// GET /api/v1/points/records?page=1&page_size=20
func (h *PointHandler) GetPointRecords(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	uid := userID.(uint)
	records, total, err := h.pointService.GetPointRecords(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取积分记录失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPointStatistics 获取积分统计
// GET /api/v1/points/statistics
func (h *PointHandler) GetPointStatistics(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	stats, err := h.pointService.GetPointStatistics(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// AwardReadPoints 奖励阅读积分（由前端调用，需要验证阅读时长）
// POST /api/v1/points/award/read
func (h *PointHandler) AwardReadPoints(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		ArticleID    uint   `json:"article_id" binding:"required"`
		ArticleTitle string `json:"article_title" binding:"required"`
		ReadDuration int    `json:"read_duration"` // 阅读时长（秒）- 可选验证
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	uid := userID.(uint)
	userPoints, record, err := h.pointService.AwardReadArticlePoints(uid, req.ArticleID, req.ArticleTitle)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "阅读积分奖励成功",
		"points_earned": record.Points,
		"user_points":   userPoints,
		"record":        record,
	})
}
