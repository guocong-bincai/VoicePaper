package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 代表一篇文章
// 对应数据库表 vp_articles
func (Article) TableName() string {
	return "vp_articles"
}

type Article struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	Title       string     `gorm:"size:255;not null" json:"title"`
	PicURL      string     `gorm:"size:512;column:pic_url" json:"pic_url"`                  // 文章封面图完整访问URL
	Pic11URL    string     `gorm:"size:512;column:pic_1_1_url" json:"pic_1_1_url"`          // 1:1比例封面图（朋友圈分享）
	Pic54URL    string     `gorm:"size:512;column:pic_5_4_url" json:"pic_5_4_url"`          // 5:4比例封面图（聊天框分享）
	Online      string     `gorm:"size:50;default:'pending';column:online" json:"online"`   // 是否上线 0:否，1:是
	CategoryID  *uint      `gorm:"index;column:category_id" json:"category_id"`             // 关联 categories.id
	PublishDate *time.Time `gorm:"type:date;index;column:publish_date" json:"publish_date"` // 发布日期（用于每日文章）
	IsDaily     bool       `gorm:"default:false;index;column:is_daily" json:"is_daily"`     // 是否为每日文章

	// URL字段（从OSS获取）
	AudioURL           string `gorm:"size:512;column:audio_url" json:"audio_url"`                       // 音频完整访问URL
	TimelineURL        string `gorm:"size:512;column:timeline_url" json:"timeline_url"`                 // 时间轴完整访问URL
	ArticleURL         string `gorm:"size:512;column:article_url" json:"article_url"`                   // 文章URL（Markdown文件）
	OriginalArticleURL string `gorm:"size:255;column:original_article_url" json:"original_article_url"` // 英文原文 OSS 地址

	// 关联
	Category  *Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Sentences []Sentence `gorm:"foreignKey:ArticleID" json:"sentences,omitempty"`
	Words     []Word     `gorm:"foreignKey:ArticleID" json:"words,omitempty"`

	// 临时字段（从ArticleURL加载的内容，不存储到数据库）
	Content string `gorm:"-" json:"content,omitempty"` // Markdown内容，从ArticleURL加载
}

// Sentence 代表文章中的一个句子 (用于听写和高亮)
// 对应数据库表 vp_sentences
func (Sentence) TableName() string {
	return "vp_sentences"
}

type Sentence struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	ArticleID   uint   `gorm:"index;not null;column:article_id" json:"article_id"`
	Text        string `gorm:"type:text;not null" json:"text"`
	Translation string `gorm:"type:text" json:"translation"` // 句子中文翻译
	Order       int    `gorm:"column:order" json:"order"`    // 排序

	// 关联
	Article Article `gorm:"foreignKey:ArticleID" json:"-"`
}
