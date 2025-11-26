package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 代表一篇文章
type Article struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Title       string `gorm:"size:255;not null" json:"title"`
	Content     string `gorm:"type:text;not null" json:"content"`       // Markdown 原文
	AudioPath   string `gorm:"size:255" json:"audio_path"`              // 本地音频路径 (相对路径)
	Status      string `gorm:"size:50;default:'pending'" json:"status"` // pending, processing, completed, failed
	ContentHash string `gorm:"size:64;index" json:"content_hash"`       // 内容哈希，用于去重

	Sentences []Sentence `gorm:"foreignKey:ArticleID" json:"sentences"`
}

// Sentence 代表文章中的一个句子 (用于听写和高亮)
type Sentence struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ArticleID uint   `gorm:"index;not null" json:"article_id"`
	Text      string `gorm:"type:text;not null" json:"text"`
	StartTime int64  `json:"start_time"` // 毫秒
	EndTime   int64  `json:"end_time"`   // 毫秒
	Order     int    `json:"order"`      // 排序
}
