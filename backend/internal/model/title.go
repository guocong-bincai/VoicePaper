package model

import (
	"time"

	"gorm.io/gorm"
)

// TitleCategory 称号分类
type TitleCategory string

const (
	TitleCategoryReading     TitleCategory = "reading"     // 阅读
	TitleCategoryDictation   TitleCategory = "dictation"   // 默写
	TitleCategoryCheckIn     TitleCategory = "check_in"    // 签到
	TitleCategoryPoints      TitleCategory = "points"      // 积分
	TitleCategorySpecial     TitleCategory = "special"     // 特殊
	TitleCategoryVocabulary  TitleCategory = "vocabulary"  // 生词本
	TitleCategoryReview      TitleCategory = "review"      // 复习
	TitleCategoryDuration    TitleCategory = "duration"    // 学习时长
	TitleCategoryAchievement TitleCategory = "achievement" // 成就
)

// TitleConditionType 称号条件类型
type TitleConditionType string

const (
	TitleConditionArticlesRead        TitleConditionType = "articles_read"        // 阅读数
	TitleConditionDictationsCompleted TitleConditionType = "dictations_completed" // 默写数
	TitleConditionContinuousCheckIns  TitleConditionType = "continuous_check_ins" // 连续签到
	TitleConditionTotalPoints         TitleConditionType = "total_points"         // 总积分
	TitleConditionCustom              TitleConditionType = "custom"               // 自定义
	TitleConditionVocabularyCount     TitleConditionType = "vocabulary_count"     // 生词本收藏数
	TitleConditionReviewCount         TitleConditionType = "review_count"         // 复习次数
	TitleConditionTotalDuration       TitleConditionType = "total_duration"       // 学习时长(分钟)
	TitleConditionPerfectStreak       TitleConditionType = "perfect_streak"       // 连续全对次数
)

// TitleRarity 称号稀有度
type TitleRarity string

const (
	TitleRarityCommon    TitleRarity = "common"    // 普通
	TitleRarityRare      TitleRarity = "rare"      // 稀有
	TitleRarityEpic      TitleRarity = "epic"      // 史诗
	TitleRarityLegendary TitleRarity = "legendary" // 传说
)

// TitleConfig 称号配置表
// 对应数据库表 vp_title_configs
func (TitleConfig) TableName() string {
	return "vp_title_configs"
}

type TitleConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	TitleKey    string        `gorm:"size:50;not null;uniqueIndex;column:title_key" json:"title_key"` // 称号唯一标识
	TitleName   string        `gorm:"size:100;not null;column:title_name" json:"title_name"`          // 称号名称
	TitleIcon   string        `gorm:"size:255;column:title_icon" json:"title_icon,omitempty"`         // 称号图标
	Description string        `gorm:"size:500;column:description" json:"description,omitempty"`       // 称号描述
	Category    TitleCategory `gorm:"size:50;not null;index;column:category" json:"category"`         // 称号分类

	// 获取条件
	ConditionType        TitleConditionType `gorm:"size:50;not null;column:condition_type" json:"condition_type"`                 // 条件类型
	ConditionValue       int                `gorm:"not null;default:0;column:condition_value" json:"condition_value"`             // 条件数值
	ConditionDescription string             `gorm:"size:255;column:condition_description" json:"condition_description,omitempty"` // 条件说明

	// 称号属性
	Rarity    TitleRarity `gorm:"size:20;default:'common';column:rarity" json:"rarity"`          // 稀有度
	SortOrder int         `gorm:"not null;default:0;index;column:sort_order" json:"sort_order"`  // 排序
	IsActive  bool        `gorm:"not null;default:true;index;column:is_active" json:"is_active"` // 是否启用
}

// UserTitle 用户称号表
// 对应数据库表 vp_user_titles
func (UserTitle) TableName() string {
	return "vp_user_titles"
}

type UserTitle struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID        uint      `gorm:"not null;uniqueIndex:uk_user_title,priority:1;index;column:user_id" json:"user_id"`
	TitleConfigID uint      `gorm:"not null;uniqueIndex:uk_user_title,priority:2;index;column:title_config_id" json:"title_config_id"`
	AwardedAt     time.Time `gorm:"not null;index;column:awarded_at" json:"awarded_at"`                 // 获得时间
	IsEquipped    bool      `gorm:"not null;default:false;index;column:is_equipped" json:"is_equipped"` // 是否佩戴

	// 关联
	User        User        `gorm:"foreignKey:UserID" json:"-"`
	TitleConfig TitleConfig `gorm:"foreignKey:TitleConfigID" json:"title_config,omitempty"`
}
