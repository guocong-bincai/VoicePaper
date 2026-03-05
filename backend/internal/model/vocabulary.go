package model

import (
	"time"

	"gorm.io/gorm"
)

// VocabularyType 生词类型
type VocabularyType string

const (
	VocabularyTypeWord     VocabularyType = "word"     // 单词
	VocabularyTypePhrase   VocabularyType = "phrase"   // 短语
	VocabularyTypeSentence VocabularyType = "sentence" // 句子
)

// ReviewType 复习方式
type ReviewType string

const (
	ReviewTypeCard      ReviewType = "card"      // 卡片
	ReviewTypeSpell     ReviewType = "spell"     // 拼写
	ReviewTypeChoice    ReviewType = "choice"    // 选择题
	ReviewTypeDictation ReviewType = "dictation" // 听写
)

// Vocabulary 生词本主表
// 对应数据库表 vp_vocabulary
func (Vocabulary) TableName() string {
	return "vp_vocabulary"
}

type Vocabulary struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	// 用户和来源
	UserID       uint   `gorm:"not null;index;column:user_id" json:"user_id"`
	ArticleID    *uint  `gorm:"index;column:article_id" json:"article_id,omitempty"`
	SentenceID   *uint  `gorm:"column:sentence_id" json:"sentence_id,omitempty"`
	WordbookID   *uint  `gorm:"index;column:wordbook_id" json:"wordbook_id,omitempty"`               // 来自哪本单词书的ID
	WordbookType string `gorm:"size:50;column:wordbook_type" json:"wordbook_type,omitempty"`         // 单词书类型：cet4/cet6/toefl等
	Source       string `gorm:"size:50;not null;default:'manual';column:source;index" json:"source"` // 来源：wordbook/article/manual

	// 生词内容
	Type               VocabularyType `gorm:"type:enum('word','phrase','sentence');not null;default:'word';column:type" json:"type"`
	Content            string         `gorm:"size:500;not null;column:content" json:"content"`
	Phonetic           string         `gorm:"size:100;column:phonetic" json:"phonetic,omitempty"`
	Meaning            string         `gorm:"type:text;column:meaning" json:"meaning,omitempty"`
	Example            string         `gorm:"type:text;column:example" json:"example,omitempty"`
	ExampleTranslation string         `gorm:"type:text;column:example_translation" json:"example_translation,omitempty"`
	Context            string         `gorm:"type:text;column:context" json:"context,omitempty"`
	Note               string         `gorm:"type:text;column:note" json:"note,omitempty"`
	Morphemes          string         `gorm:"type:text;column:morphemes" json:"morphemes,omitempty"` // 词根词缀分析结果，JSON格式

	// 学习状态（SM-2算法）
	MasteryLevel int        `gorm:"default:0;column:mastery_level" json:"mastery_level"`                  // 掌握等级 0-5
	EaseFactor   float64    `gorm:"type:decimal(4,2);default:2.50;column:ease_factor" json:"ease_factor"` // SM-2易度因子
	IntervalDays int        `gorm:"default:0;column:interval_days" json:"interval_days"`                  // 当前复习间隔（天）
	Repetitions  int        `gorm:"default:0;column:repetitions" json:"repetitions"`                      // 连续正确次数
	NextReviewAt *time.Time `gorm:"column:next_review_at" json:"next_review_at,omitempty"`
	LastReviewAt *time.Time `gorm:"column:last_review_at" json:"last_review_at,omitempty"`

	// 统计
	ReviewCount  int `gorm:"default:0;column:review_count" json:"review_count"`
	CorrectCount int `gorm:"default:0;column:correct_count" json:"correct_count"`
	WrongCount   int `gorm:"default:0;column:wrong_count" json:"wrong_count"`

	// 分类标签
	Tags      string `gorm:"size:500;column:tags" json:"tags,omitempty"` // JSON数组格式
	IsStarred bool   `gorm:"default:false;column:is_starred" json:"is_starred"`

	// 关联
	User    User     `gorm:"foreignKey:UserID" json:"-"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// VocabularyReview 复习记录表
// 对应数据库表 vp_vocabulary_reviews
func (VocabularyReview) TableName() string {
	return "vp_vocabulary_reviews"
}

type VocabularyReview struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	UserID       uint `gorm:"not null;index;column:user_id" json:"user_id"`
	VocabularyID uint `gorm:"not null;index;column:vocabulary_id" json:"vocabulary_id"`

	// 复习详情
	ReviewType     ReviewType `gorm:"type:enum('card','spell','choice','dictation');not null;default:'card';column:review_type" json:"review_type"`
	Quality        int        `gorm:"not null;column:quality" json:"quality"` // SM-2评分 0-5
	IsCorrect      bool       `gorm:"not null;column:is_correct" json:"is_correct"`
	ResponseTimeMs *int       `gorm:"column:response_time_ms" json:"response_time_ms,omitempty"`
	UserAnswer     string     `gorm:"type:text;column:user_answer" json:"user_answer,omitempty"`

	// 复习前后状态变化
	PrevInterval   *int     `gorm:"column:prev_interval" json:"prev_interval,omitempty"`
	NewInterval    *int     `gorm:"column:new_interval" json:"new_interval,omitempty"`
	PrevEaseFactor *float64 `gorm:"type:decimal(4,2);column:prev_ease_factor" json:"prev_ease_factor,omitempty"`
	NewEaseFactor  *float64 `gorm:"type:decimal(4,2);column:new_ease_factor" json:"new_ease_factor,omitempty"`

	// 关联
	User       User       `gorm:"foreignKey:UserID" json:"-"`
	Vocabulary Vocabulary `gorm:"foreignKey:VocabularyID" json:"-"`
}

// VocabularyFolder 生词本文件夹
// 对应数据库表 vp_vocabulary_folders
func (VocabularyFolder) TableName() string {
	return "vp_vocabulary_folders"
}

type VocabularyFolder struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID      uint   `gorm:"not null;index;column:user_id" json:"user_id"`
	Name        string `gorm:"size:100;not null;column:name" json:"name"`
	Description string `gorm:"size:500;column:description" json:"description,omitempty"`
	Color       string `gorm:"size:20;default:'#3b82f6';column:color" json:"color"`
	Icon        string `gorm:"size:50;column:icon" json:"icon,omitempty"`
	SortOrder   int    `gorm:"default:0;column:sort_order" json:"sort_order"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// VocabularyFolderItem 生词-文件夹关联
// 对应数据库表 vp_vocabulary_folder_items
func (VocabularyFolderItem) TableName() string {
	return "vp_vocabulary_folder_items"
}

type VocabularyFolderItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	FolderID     uint      `gorm:"not null;index;column:folder_id" json:"folder_id"`
	VocabularyID uint      `gorm:"not null;index;column:vocabulary_id" json:"vocabulary_id"`

	// 关联
	Folder     VocabularyFolder `gorm:"foreignKey:FolderID" json:"-"`
	Vocabulary Vocabulary       `gorm:"foreignKey:VocabularyID" json:"-"`
}

// VocabularyDailyStats 学习统计日报
// 对应数据库表 vp_vocabulary_daily_stats
func (VocabularyDailyStats) TableName() string {
	return "vp_vocabulary_daily_stats"
}

type VocabularyDailyStats struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	UserID   uint      `gorm:"not null;index;column:user_id" json:"user_id"`
	StatDate time.Time `gorm:"type:date;not null;column:stat_date" json:"stat_date"`

	// 当日统计
	NewWords          int      `gorm:"default:0;column:new_words" json:"new_words"`
	ReviewedWords     int      `gorm:"default:0;column:reviewed_words" json:"reviewed_words"`
	MasteredWords     int      `gorm:"default:0;column:mastered_words" json:"mastered_words"`
	ReviewTimeSeconds int      `gorm:"default:0;column:review_time_seconds" json:"review_time_seconds"`
	CorrectRate       *float64 `gorm:"type:decimal(5,2);column:correct_rate" json:"correct_rate,omitempty"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}
