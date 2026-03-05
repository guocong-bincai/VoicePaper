package model

import (
	"time"

	"gorm.io/gorm"
)

// UserDailyStats 用户每日统计表
// 用于记录小程序端的每日学习数据
// 对应数据库表 vp_user_daily_stats
func (UserDailyStats) TableName() string {
	return "vp_user_daily_stats"
}

type UserDailyStats struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID   uint      `gorm:"not null;uniqueIndex:idx_user_date;column:user_id" json:"user_id"`
	StatDate time.Time `gorm:"type:date;not null;uniqueIndex:idx_user_date;column:stat_date" json:"stat_date"`

	// 时长统计（秒）
	TotalDurationSeconds int `gorm:"default:0;column:total_duration_seconds" json:"total_duration_seconds"` // 今日总时长（秒）

	// 单词学习统计
	NewWords      int `gorm:"default:0;column:new_words" json:"new_words"`           // 新学单词数
	ReviewedWords int `gorm:"default:0;column:reviewed_words" json:"reviewed_words"` // 复习单词数
	CorrectCount  int `gorm:"default:0;column:correct_count" json:"correct_count"`   // 正确次数
	TotalAttempts int `gorm:"default:0;column:total_attempts" json:"total_attempts"` // 总尝试次数

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// GetCorrectRate 计算正确率
func (s *UserDailyStats) GetCorrectRate() float64 {
	if s.TotalAttempts == 0 {
		return 0
	}
	return float64(s.CorrectCount) / float64(s.TotalAttempts) * 100
}
