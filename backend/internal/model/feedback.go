package model

import (
	"time"

	"gorm.io/gorm"
)

// Feedback 用户反馈模型
type Feedback struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Type        string         `gorm:"type:varchar(20);not null;index" json:"type"` // feature, bug, ui, other
	Description string         `gorm:"type:text;not null" json:"description"`
	Contact     string         `gorm:"type:varchar(200)" json:"contact,omitempty"`       // 可选联系方式
	UserID      *uint          `gorm:"index" json:"user_id,omitempty"`                   // 可选，关联用户
	Status      string         `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, processing, resolved, closed
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Feedback) TableName() string {
	return "vp_feedbacks"
}
