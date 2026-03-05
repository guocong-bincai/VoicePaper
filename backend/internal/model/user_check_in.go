package model

import (
	"time"

	"gorm.io/gorm"
)

// UserCheckIn 用户签到记录表
// 对应数据库表 vp_user_check_ins
func (UserCheckIn) TableName() string {
	return "vp_user_check_ins"
}

type UserCheckIn struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"` // 签到时间
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	UserID         uint      `gorm:"not null;uniqueIndex:uk_user_date,priority:1;index;column:user_id" json:"user_id"`
	CheckInDate    time.Time `gorm:"type:date;not null;uniqueIndex:uk_user_date,priority:2;index;column:check_in_date" json:"check_in_date"` // 签到日期
	PointsAwarded  int       `gorm:"not null;default:0;column:points_awarded" json:"points_awarded"`                                         // 本次签到获得的积分
	ContinuousDays int       `gorm:"not null;default:1;column:continuous_days" json:"continuous_days"`                                       // 签到时的连续天数
	IsMakeup       bool      `gorm:"not null;default:false;column:is_makeup" json:"is_makeup"`                                               // 是否补签

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}
