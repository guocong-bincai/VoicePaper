package model

import (
	"time"
)

// DictationType 听写类型
type DictationType string

const (
	DictationTypeSentence DictationType = "sentence" // 句子听写
	DictationTypeWord     DictationType = "word"     // 单词听写
)

// DictationRecord 听写记录表（记录学生的学习进度）
type DictationRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 用户信息（关联用户表）
	UserID *uint `gorm:"index;column:user_id" json:"user_id,omitempty"` // 关联users.id（使用指针允许NULL，支持未登录用户）

	ArticleID uint `gorm:"index;not null" json:"article_id"`

	// 听写类型和关联
	DictationType DictationType `gorm:"size:20;not null;index" json:"dictation_type"` // 'sentence' | 'word'
	SentenceID    *uint         `gorm:"index" json:"sentence_id"`                     // 如果是句子听写，关联sentences.id
	WordID        *uint         `gorm:"index" json:"word_id"`                         // 如果是单词听写，关联words.id

	// 听写结果
	UserAnswer string `gorm:"type:text" json:"user_answer"`          // 用户输入的答案
	IsCorrect  bool   `gorm:"default:false;index" json:"is_correct"` // 是否正确
	Score      int    `gorm:"default:0" json:"score"`                // 得分（0-100）

	// 学习数据
	AttemptCount int       `gorm:"default:1" json:"attempt_count"` // 尝试次数
	TimeSpent    int       `gorm:"default:0" json:"time_spent"`    // 花费时间（秒）
	LastAttempt  time.Time `json:"last_attempt"`                   // 最后尝试时间

	// 关联
	Article  Article   `gorm:"foreignKey:ArticleID" json:"-"`
	Sentence *Sentence `gorm:"foreignKey:SentenceID" json:"-"`
	Word     *Word     `gorm:"foreignKey:WordID" json:"-"`
}

// TableName 指定表名
func (DictationRecord) TableName() string {
	return "vp_dictation_records"
}
