package model

import (
	"time"

	"gorm.io/gorm"
)

// AppConfig 应用动态配置表
// 用于控制小程序审核模式、文章可见范围等
type AppConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	ConfigKey string `gorm:"uniqueIndex;size:50;not null" json:"config_key"` // 配置键，例如 "miniprogram_audit"

	// 审核模式配置
	IsAuditMode bool   `gorm:"default:false" json:"is_audit_mode"` // 是否开启审核模式
	StartDate   string `gorm:"size:20" json:"start_date"`          // 可见文章起始日期 (YYYY-MM-DD)，为空则不限制
	EndDate     string `gorm:"size:20" json:"end_date"`            // 可见文章结束日期 (YYYY-MM-DD)，为空则不限制

	Description string `gorm:"size:255" json:"description"` // 描述
}

func (AppConfig) TableName() string {
	return "vp_app_configs"
}
