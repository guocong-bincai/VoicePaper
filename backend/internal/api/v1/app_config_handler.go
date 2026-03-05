package v1

import (
	"net/http"
	"voicepaper/internal/repository"

	"github.com/gin-gonic/gin"
)

type AppConfigHandler struct {
	repo *repository.AppConfigRepository
}

func NewAppConfigHandler() *AppConfigHandler {
	return &AppConfigHandler{
		repo: repository.NewAppConfigRepository(),
	}
}

// GetAuditConfig 获取小程序审核配置
// GET /api/v1/config/audit
func (h *AppConfigHandler) GetAuditConfig(c *gin.Context) {
	config, err := h.repo.GetConfigByKey("miniprogram_audit")
	if err != nil {
		// 如果不存在，返回默认关闭
		c.JSON(http.StatusOK, gin.H{
			"is_audit_mode": false,
			"start_date":    "2026-01-20",
			"end_date":      "",
			"error":         "Config not found, using default",
		})
		return
	}
	c.JSON(http.StatusOK, config)
}
