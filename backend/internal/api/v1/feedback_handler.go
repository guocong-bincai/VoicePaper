package v1

import (
	"log"
	"net/http"
	"voicepaper/internal/model"
	"voicepaper/internal/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FeedbackHandler 反馈处理器
type FeedbackHandler struct {
	repo *repository.FeedbackRepository
}

// NewFeedbackHandler 创建反馈处理器实例
func NewFeedbackHandler(db *gorm.DB) *FeedbackHandler {
	return &FeedbackHandler{
		repo: repository.NewFeedbackRepository(db),
	}
}

// SubmitFeedback 提交反馈（可选认证）
// POST /api/v1/feedback
func (h *FeedbackHandler) SubmitFeedback(c *gin.Context) {
	var req struct {
		Type        string `json:"type" binding:"required,oneof=feature bug ui other"`
		Description string `json:"description" binding:"required"`
		Contact     string `json:"contact"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建反馈记录
	feedback := &model.Feedback{
		Type:        req.Type,
		Description: req.Description,
		Contact:     req.Contact,
		Status:      "pending",
	}

	// 如果用户已登录，关联用户ID（可选）
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			feedback.UserID = &uid
		}
	}

	if err := h.repo.Create(feedback); err != nil {
		log.Printf("❌ 创建反馈失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交失败"})
		return
	}

	log.Printf("✅ 收到新反馈 [%s]: %s", req.Type, req.Description[:min(50, len(req.Description))])

	c.JSON(http.StatusOK, gin.H{
		"message": "感谢您的反馈！",
		"id":      feedback.ID,
	})
}

// min 辅助函数：返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
