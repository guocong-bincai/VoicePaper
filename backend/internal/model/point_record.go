package model

import (
	"time"

	"gorm.io/gorm"
)

// PointType 积分类型
type PointType string

const (
	PointTypeReadArticle       PointType = "read_article"        // 阅读文章
	PointTypeWordDictation     PointType = "word_dictation"      // 单词默写
	PointTypeSentenceDictation PointType = "sentence_dictation"  // 句子默写
	PointTypeDailyCheckIn      PointType = "daily_check_in"      // 每日签到
	PointTypeContinuousCheckIn PointType = "continuous_check_in" // 连续签到奖励
	PointTypeCompleteArticle   PointType = "complete_article"    // 完成整篇文章默写
	PointTypeVocabularyReview  PointType = "vocabulary_review"   // 生词复习
	PointTypeWordbookStudy     PointType = "wordbook_study"      // 单词书学习
	PointTypeAdminAdjust       PointType = "admin_adjust"        // 管理员调整
)

// PointRecord 积分记录表
// 对应数据库表 vp_point_records
func (PointRecord) TableName() string {
	return "vp_point_records"
}

type PointRecord struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID      uint      `gorm:"not null;index;column:user_id" json:"user_id"`
	Points      int       `gorm:"not null;column:points" json:"points"`                     // 积分变动数量（正数为增加，负数为扣除）
	Type        PointType `gorm:"size:50;not null;index;column:type" json:"type"`           // 积分类型
	Description string    `gorm:"size:255;column:description" json:"description,omitempty"` // 积分说明

	// 关联字段（可选）
	ArticleID         *uint `gorm:"index;column:article_id" json:"article_id,omitempty"`             // 关联文章ID
	DictationRecordID *uint `gorm:"column:dictation_record_id" json:"dictation_record_id,omitempty"` // 关联默写记录ID
	CheckInID         *uint `gorm:"column:check_in_id" json:"check_in_id,omitempty"`                 // 关联签到记录ID

	// 余额快照
	BalanceBefore int `gorm:"not null;default:0;column:balance_before" json:"balance_before"` // 变动前积分余额
	BalanceAfter  int `gorm:"not null;default:0;column:balance_after" json:"balance_after"`   // 变动后积分余额

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// PointRewardConfig 积分奖励配置
var PointRewardConfig = map[PointType]int{
	PointTypeReadArticle:       5,  // 阅读文章 +5
	PointTypeWordDictation:     2,  // 单词默写正确 +2
	PointTypeSentenceDictation: 5,  // 句子默写正确 +5
	PointTypeDailyCheckIn:      10, // 每日签到 +10
	PointTypeCompleteArticle:   20, // 完成整篇默写 +20
}

// WordbookStudyPointsConfig 单词书学习积分配置
// 不认识: +1, 模糊: +2, 认识: +3
var WordbookStudyPointsConfig = map[string]int{
	"forget": 1, // 不认识 +1
	"fuzzy":  2, // 模糊 +2
	"know":   3, // 认识 +3
}

// ContinuousCheckInRewards 连续签到奖励配置
var ContinuousCheckInRewards = map[int]int{
	3:  30,  // 连续3天 +30
	7:  100, // 连续7天 +100
	30: 500, // 连续30天 +500
}
