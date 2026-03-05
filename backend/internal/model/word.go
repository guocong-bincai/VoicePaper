package model

import (
	"time"

	"gorm.io/gorm"
)

// Word 单词表（用于听写练习）
// 对应数据库表 vp_words
func (Word) TableName() string {
	return "vp_words"
}

type Word struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	ArticleID          uint   `gorm:"index;not null;column:article_id" json:"article_id"`
	Text               string `gorm:"size:100;not null;index" json:"text"`  // 单词文本
	Phonetic           string `gorm:"size:100" json:"phonetic"`             // 音标，如 /wɜːld/
	Meaning            string `gorm:"size:255" json:"meaning"`              // 中文释义
	Example            string `gorm:"type:text" json:"example"`             // 例句（4级/6级/考研真题）
	ExampleTranslation string `gorm:"type:text" json:"example_translation"` // 例句中文翻译

	// 单词属性
	Level     int  `gorm:"default:0;column:level" json:"level"`                       // 难度等级 0-5
	Frequency int  `gorm:"default:0;column:frequency" json:"frequency"`               // 词频（可用于推荐）
	Order     int  `gorm:"not null;column:order" json:"order"`                        // 排序（在文章中的出现顺序）
	IsKeyWord bool `gorm:"default:false;index;column:is_key_word" json:"is_key_word"` // 是否为重点单词

	// 关联
	Article Article `gorm:"foreignKey:ArticleID" json:"-"`
}
