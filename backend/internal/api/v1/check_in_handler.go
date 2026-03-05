package v1

import (
	"net/http"
	"strconv"
	"time"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
)

type CheckInHandler struct {
	checkInService *service.CheckInService
}

func NewCheckInHandler() *CheckInHandler {
	return &CheckInHandler{
		checkInService: service.NewCheckInService(repository.DB),
	}
}

// CheckIn 每日签到
// POST /api/v1/check-in
func (h *CheckInHandler) CheckIn(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	result, err := h.checkInService.CheckIn(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCheckInStatus 获取签到状态
// GET /api/v1/check-in/status
func (h *CheckInHandler) GetCheckInStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	status, err := h.checkInService.GetCheckInStatus(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取签到状态失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetCheckInCalendar 获取签到日历
// GET /api/v1/check-in/calendar?year=2025&month=12
func (h *CheckInHandler) GetCheckInCalendar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 获取年月参数（默认当前月）
	now := time.Now()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))

	// 参数验证
	if year < 2020 || year > 2100 {
		year = now.Year()
	}
	if month < 1 || month > 12 {
		month = int(now.Month())
	}

	uid := userID.(uint)
	calendar, err := h.checkInService.GetCheckInCalendar(uid, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取签到日历失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

// GetCheckInRanking 获取签到排行榜
// GET /api/v1/check-in/ranking?limit=20&sort_by=total
func (h *CheckInHandler) GetCheckInRanking(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort_by", "total")

	uid := userID.(uint)
	ranking, err := h.checkInService.GetCheckInRanking(limit, sortBy, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取签到排行榜失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ranking)
}

// GetMakeupCardInfo 获取补签卡信息
// GET /api/v1/check-in/makeup-card
func (h *CheckInHandler) GetMakeupCardInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	info, err := h.checkInService.GetMakeupCardInfo(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取补签卡信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// BuyMakeupCard 购买补签卡
// POST /api/v1/check-in/makeup-card/buy
func (h *CheckInHandler) BuyMakeupCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Count = 1 // 默认1张
	}

	uid := userID.(uint)
	result, err := h.checkInService.BuyMakeupCard(uid, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UseMakeupCard 使用补签卡补签
// POST /api/v1/check-in/makeup-card/use
func (h *CheckInHandler) UseMakeupCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		Date string `json:"date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供补签日期"})
		return
	}

	uid := userID.(uint)
	result, err := h.checkInService.UseMakeupCard(uid, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
