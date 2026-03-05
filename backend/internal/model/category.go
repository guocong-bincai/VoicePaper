package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类表
// 对应数据库表 vp_categories
func (Category) TableName() string {
	return "vp_categories"
}

type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`

	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`        // 分类名称
	Description string `gorm:"type:text" json:"description"`                     // 分类描述
	Icon        string `gorm:"size:255" json:"icon"`                              // 分类图标
	Sort        int    `gorm:"default:0;index;column:sort" json:"sort"`          // 排序
	IsActive    bool   `gorm:"default:true;index;column:is_active" json:"is_active"` // 是否启用

	// 关联
	Articles []Article `gorm:"foreignKey:CategoryID" json:"-"`
}

