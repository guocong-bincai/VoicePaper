package model

import (
	"time"

	"gorm.io/gorm"
)

// Wordbook 单词书表
type Wordbook struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Word               string         `gorm:"size:100;not null" json:"word"`
	Phonetic           string         `gorm:"size:100" json:"phonetic"`
	WordType           string         `gorm:"size:20;not null" json:"word_type"` // junior/senior/cet4/cet6/postgrad/toefl
	Frequency          uint           `gorm:"default:0" json:"frequency"`
	Difficulty         uint8          `gorm:"default:3" json:"difficulty"`
	Source             string         `gorm:"size:50" json:"source"`
	Meaning            string         `gorm:"type:text" json:"meaning"`
	MeaningAnalysis    string         `gorm:"type:text" json:"meaning_analysis"`
	Example            string         `gorm:"type:text" json:"example"`
	ExampleTranslation string         `gorm:"type:text" json:"example_translation"`
	ExamplesJSON       string         `gorm:"column:examples_json;type:json" json:"examples_json"`
	Phrases            string         `gorm:"type:json" json:"phrases"`
	Root               string         `gorm:"size:100" json:"root"`
	RootAnalysis       string         `gorm:"type:text" json:"root_analysis"`
	Affix              string         `gorm:"size:100" json:"affix"`
	AffixAnalysis      string         `gorm:"type:text" json:"affix_analysis"`
	Etymology          string         `gorm:"type:text" json:"etymology"`
	CulturalBackground string         `gorm:"type:text" json:"cultural_background"`
	WordForms          string         `gorm:"column:word_forms;type:json" json:"word_forms"`
	MemoryTips         string         `gorm:"type:text" json:"memory_tips"`
	StoryEn            string         `gorm:"column:story_en;type:text" json:"story_en"`
	StoryCn            string         `gorm:"column:story_cn;type:text" json:"story_cn"`
	DrawExplain        string         `gorm:"type:text" json:"draw_explain"`
	DrawPrompt         string         `gorm:"size:500" json:"draw_prompt"`
	ImageURL           string         `gorm:"column:image_url;size:500" json:"image_url"`
	AnalysisFull       string         `gorm:"type:text" json:"analysis_full"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Wordbook) TableName() string {
	return "vp_wordbook"
}

// WordbookInfo 单词书基础信息
type WordbookInfo struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Type        string         `gorm:"size:20;not null;uniqueIndex" json:"type"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	CoverURL    string         `gorm:"column:cover_url;size:500" json:"cover_url"`
	Hotness     uint           `gorm:"default:0" json:"hotness"`
	StudyCount  uint           `gorm:"default:0" json:"study_count"`
	WordCount   uint           `gorm:"default:0" json:"word_count"`
	Category    string         `gorm:"size:50" json:"category"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WordbookInfo) TableName() string {
	return "vp_wordbook_info"
}

// WordbookProgress 单词书学习进度
type WordbookProgress struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index:idx_user_wordtype" json:"user_id"`
	WordType     string         `gorm:"size:20;not null;index:idx_user_wordtype" json:"word_type"`
	CurrentIndex int            `gorm:"default:0" json:"current_index"` // 当前学习到的索引位置
	TotalWords   int            `gorm:"default:0" json:"total_words"`   // 该词书总单词数
	LastWordID   uint           `json:"last_word_id"`                   // 最后学习的单词ID
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WordbookProgress) TableName() string {
	return "vp_wordbook_progress"
}

// WordbookUserOrder 用户单词书学习序列（支持顺序/乱序）
type WordbookUserOrder struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;uniqueIndex:uk_user_word_type" json:"user_id"`
	WordType     string         `gorm:"size:20;not null;uniqueIndex:uk_user_word_type" json:"word_type"`
	IsRandom     bool           `gorm:"default:false" json:"is_random"` // 是否乱序模式
	WordSequence string         `gorm:"type:text" json:"word_sequence"` // JSON格式的单词ID序列
	CurrentIndex int            `gorm:"default:0" json:"current_index"` // 当前学习位置
	TotalWords   int            `gorm:"default:0" json:"total_words"`   // 总单词数
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WordbookUserOrder) TableName() string {
	return "vp_wordbook_user_order"
}

// WordbookBook 单词与单词书关联表
type WordbookBook struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WordID    uint      `gorm:"not null;index:uk_word_book" json:"word_id"`
	BookType  string    `gorm:"size:50;not null;index:uk_word_book" json:"book_type"`
	CreatedAt time.Time `json:"created_at"`
}

func (WordbookBook) TableName() string {
	return "vp_wordbook_books"
}
