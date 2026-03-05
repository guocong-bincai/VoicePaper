package v1

import (
	"net/http"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"
	"voicepaper/internal/service"

	"github.com/gin-gonic/gin"
)

type TitleHandler struct {
	titleService *service.TitleService
}

func NewTitleHandler() *TitleHandler {
	return &TitleHandler{
		titleService: service.NewTitleService(repository.DB),
	}
}

// GetAllTitles 获取所有可用称号
// GET /api/v1/titles/all
func (h *TitleHandler) GetAllTitles(c *gin.Context) {
	titles, err := h.titleService.GetAllTitles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取称号列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, titles)
}

// GetMyTitles 获取我的称号
// GET /api/v1/titles/me
func (h *TitleHandler) GetMyTitles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	titles, err := h.titleService.GetUserTitles(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取我的称号失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, titles)
}

// GetTitleProgress 获取称号进度
// GET /api/v1/titles/progress
func (h *TitleHandler) GetTitleProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	progress, err := h.titleService.GetTitleProgress(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取称号进度失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// EquipTitle 佩戴称号
// POST /api/v1/titles/equip
func (h *TitleHandler) EquipTitle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var req struct {
		TitleConfigID uint `json:"title_config_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	uid := userID.(uint)
	err := h.titleService.EquipTitle(uid, req.TitleConfigID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "称号佩戴成功"})
}

// UnequipTitle 取消佩戴称号
// POST /api/v1/titles/unequip
func (h *TitleHandler) UnequipTitle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	err := h.titleService.UnequipTitle(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消佩戴失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已取消佩戴"})
}

// GetEquippedTitle 获取当前佩戴的称号
// GET /api/v1/titles/equipped
func (h *TitleHandler) GetEquippedTitle(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(uint)
	title, err := h.titleService.GetEquippedTitle(uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "未佩戴称号"})
		return
	}

	c.JSON(http.StatusOK, title)
}

// GetTitlesByCategory 根据分类获取称号
// GET /api/v1/titles/category/:category
func (h *TitleHandler) GetTitlesByCategory(c *gin.Context) {
	category := c.Param("category")

	titles, err := h.titleService.GetTitlesByCategory(model.TitleCategory(category))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取称号列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, titles)
}
