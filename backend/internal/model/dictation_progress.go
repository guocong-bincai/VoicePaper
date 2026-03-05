package model

import (
	"time"
)

// UserDictationProgress 用户默写练习进度
type UserDictationProgress struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index:idx_progress_user_article_type,priority:1" json:"user_id"`
	ArticleID      uint      `gorm:"not null;index:idx_progress_user_article_type,priority:2" json:"article_id"`
	DictationType  string    `gorm:"type:varchar(20);not null;index:idx_progress_user_article_type,priority:3" json:"dictation_type"` // word/sentence
	CurrentIndex   int       `gorm:"not null;default:0" json:"current_index"`                                                         // 当前索引位置
	TotalItems     int       `gorm:"not null;default:0" json:"total_items"`                                                           // 总数量
	Score          int       `gorm:"not null;default:0" json:"score"`                                                                 // 得分
	Completed      bool      `gorm:"not null;default:false;index" json:"completed"`                                                   // 是否完成
	LastPracticeAt time.Time `gorm:"not null;index" json:"last_practice_at"`                                                          // 最后练习时间
	CreatedAt      time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null" json:"updated_at"`
}

// TableName 指定表名
func (UserDictationProgress) TableName() string {
	return "vp_user_dictation_progress"
}
