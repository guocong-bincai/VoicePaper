package model

import (
	"time"

	"gorm.io/gorm"
)

// ReadingProgress 用户阅读进度
// 对应数据库表 vp_reading_progress
func (ReadingProgress) TableName() string {
	return "vp_reading_progress"
}

type ReadingProgress struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID      uint      `gorm:"uniqueIndex:idx_user_article;not null;column:user_id" json:"user_id"`
	ArticleID   uint      `gorm:"uniqueIndex:idx_user_article;not null;column:article_id" json:"article_id"`
	CurrentTime float64   `gorm:"not null;default:0;column:current_time" json:"current_time"`     // 当前播放时间（秒）
	Duration    float64   `gorm:"not null;default:0;column:duration" json:"duration"`             // 音频总时长（秒）
	Progress    float64   `gorm:"not null;default:0;column:progress" json:"progress"`             // 阅读进度百分比（0-100）
	IsCompleted bool      `gorm:"not null;default:false;column:is_completed" json:"is_completed"` // 是否已完成阅读（进度>=80%）
	ReadCount   int       `gorm:"not null;default:0;column:read_count" json:"read_count"`         // 阅读次数
	TotalTime   int       `gorm:"not null;default:0;column:total_time" json:"total_time"`         // 累计阅读时间（秒）
	LastReadAt  time.Time `gorm:"column:last_read_at" json:"last_read_at"`                        // 最后阅读时间

	// 关联
	Article Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}
