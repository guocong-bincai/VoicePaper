package model

import (
	"time"

	"gorm.io/gorm"
)

// UserPoints 用户积分表
// 对应数据库表 vp_user_points
func (UserPoints) TableName() string {
	return "vp_user_points"
}

type UserPoints struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID        uint   `gorm:"not null;uniqueIndex;column:user_id" json:"user_id"`
	TotalPoints   int    `gorm:"not null;default:0;column:total_points" json:"total_points"`     // 累计总积分
	CurrentPoints int    `gorm:"not null;default:0;column:current_points" json:"current_points"` // 当前可用积分
	Level         int    `gorm:"not null;default:1;index;column:level" json:"level"`             // 用户等级
	LevelName     string `gorm:"size:50;default:'初学者';column:level_name" json:"level_name"`      // 等级名称

	// 统计字段
	TotalArticlesRead        int   `gorm:"not null;default:0;column:total_articles_read" json:"total_articles_read"`               // 累计阅读文章数
	TotalDictationsCompleted int   `gorm:"not null;default:0;column:total_dictations_completed" json:"total_dictations_completed"` // 累计完成默写次数
	TotalCheckIns            int   `gorm:"not null;default:0;column:total_check_ins" json:"total_check_ins"`                       // 累计签到天数
	ContinuousCheckIns       int   `gorm:"not null;default:0;column:continuous_check_ins" json:"continuous_check_ins"`             // 当前连续签到天数
	MaxContinuousCheckIns    int   `gorm:"not null;default:0;column:max_continuous_check_ins" json:"max_continuous_check_ins"`     // 最大连续签到天数
	MakeupCards              int   `gorm:"not null;default:0;column:makeup_cards" json:"makeup_cards"`                             // 拥有的补签卡数量
	TotalDurationMinutes     int64 `gorm:"not null;default:0;column:total_duration_minutes" json:"total_duration_minutes"`         // 累积学习时长（分钟）

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// CalculateLevel 根据总积分计算等级
func (up *UserPoints) CalculateLevel() {
	switch {
	case up.TotalPoints >= 50000:
		up.Level = 10
		up.LevelName = "传奇大师"
	case up.TotalPoints >= 30000:
		up.Level = 9
		up.LevelName = "至尊宗师"
	case up.TotalPoints >= 20000:
		up.Level = 8
		up.LevelName = "超凡大师"
	case up.TotalPoints >= 10000:
		up.Level = 7
		up.LevelName = "精英专家"
	case up.TotalPoints >= 5000:
		up.Level = 6
		up.LevelName = "资深学者"
	case up.TotalPoints >= 3000:
		up.Level = 5
		up.LevelName = "进阶学者"
	case up.TotalPoints >= 1000:
		up.Level = 4
		up.LevelName = "中级学习者"
	case up.TotalPoints >= 500:
		up.Level = 3
		up.LevelName = "初级学习者"
	case up.TotalPoints >= 100:
		up.Level = 2
		up.LevelName = "入门新手"
	default:
		up.Level = 1
		up.LevelName = "初学者"
	}
}
