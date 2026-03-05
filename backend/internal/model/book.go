package model

import (
	"time"
)

// BookInfo 书籍详情表
// 对应数据库表 vp_book_info
func (BookInfo) TableName() string {
	return "vp_book_info"
}

type BookInfo struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	BookID      int    `gorm:"column:book_id;uniqueIndex:idx_book_id;not null;default:0" json:"book_id"` // 图书ID
	Name        string `gorm:"column:name;size:256;not null;default:''" json:"name"`                     // 书名
	Cover       string `gorm:"column:cover;size:256;not null;default:''" json:"cover"`                   // 封面图片
	SmallPic    string `gorm:"column:small_pic;size:256;not null;default:''" json:"small_pic"`           // 小图
	Subtitle    string `gorm:"column:subtitle;size:256;not null;default:''" json:"subtitle"`             // 副标题
	AuthorName  string `gorm:"column:author_name;size:256;not null;default:''" json:"author_name"`       // 作者
	Value       string `gorm:"column:value;size:1024;not null;default:''" json:"value"`                  // 描述
	AuthorDesc  string `gorm:"column:author_desc;size:2048;not null;default:''" json:"author_desc"`      // 作者介绍
	WhatInside  string `gorm:"column:what_inside;size:2048;not null;default:''" json:"what_inside"`      // 内容介绍
	Learn       string `gorm:"column:learn;size:4096;not null;default:''" json:"learn"`                  // 学习的意义
	Duration    int    `gorm:"column:duration;not null;default:0" json:"duration"`                       // 音频总时长(s)
	TotalSize   int    `gorm:"column:total_size;not null;default:0" json:"total_size"`                   // 音频大小(byte)
	VoiceName   string `gorm:"column:voice_name;size:128;not null;default:''" json:"voice_name"`         // 音色
	VoiceIDs    string `gorm:"column:voice_ids;size:128;not null;default:''" json:"voice_ids"`           // 声音id
	VoiceSource int8   `gorm:"column:voice_source;not null;default:1" json:"voice_source"`               // 1-微软 2-OpenAI
	PointCount  int    `gorm:"column:point_count;not null;default:0" json:"point_count"`                 // 观点数
	KeyPoints   string `gorm:"column:key_points;type:text" json:"key_points"`                            // 关键观点
	BookType    int8   `gorm:"column:book_type;not null;default:0" json:"book_type"`                     // 0-默认类型 1-小说 2-传记
	IsOnline    int8   `gorm:"column:is_online;not null;default:0" json:"is_online"`                     // 0-未上线 1-已上线 2-待确认
}

// BookPoint 书籍内容表
// 对应数据库表 vp_book_points
func (BookPoint) TableName() string {
	return "vp_book_points"
}

type BookPoint struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	BookID       int    `gorm:"column:book_id;index:idx_book_id;uniqueIndex:idx_book_point_id;not null;default:0" json:"book_id"` // 图书ID
	Name         string `gorm:"column:name;size:256;not null;default:''" json:"name"`                                             // 书名
	Subtitle     string `gorm:"column:subtitle;size:256;not null;default:''" json:"subtitle"`                                     // 副标题
	AuthorName   string `gorm:"column:author_name;size:256;not null;default:''" json:"author_name"`                               // 作者
	PointID      int8   `gorm:"column:point_id;uniqueIndex:idx_book_point_id;not null;default:0" json:"point_id"`                 // 内容id
	PointTitle   string `gorm:"column:point_title;size:512;not null;default:''" json:"point_title"`                               // 内容标题
	PointInfo    string `gorm:"column:point_info;type:text" json:"point_info"`                                                    // 内容信息
	LrcPointInfo string `gorm:"column:lrc_point_info;type:text" json:"lrc_point_info"`                                            // lrc内容信息
	AudioURL     string `gorm:"column:audio_url;size:256;not null;default:''" json:"audio_url"`                                   // 音频文件
	AudioTimes   int    `gorm:"column:audio_times;not null;default:0" json:"audio_times"`                                         // 音频时长(s)
	AudioSize    int    `gorm:"column:audio_size;not null;default:0" json:"audio_size"`                                           // 音频大小(byte)
	AudioMD5     string `gorm:"column:audio_md5;size:64;not null;default:''" json:"audio_md5"`                                    // 音频md5
	VoiceName    string `gorm:"column:voice_name;size:128;not null;default:''" json:"voice_name"`                                 // 音色
	VoiceSource  int8   `gorm:"column:voice_source;not null;default:1" json:"voice_source"`                                       // 1-微软 2-OpenAI
	BookType     int8   `gorm:"column:book_type;not null;default:0" json:"book_type"`                                             // 书籍类型：0-默认 1-小说 2-人物传记
	Status       int8   `gorm:"column:status;not null;default:0" json:"status"`                                                   // 数据状态: 1-数据录入 2-审核通过 3-审核未通过
	IsOnline     int8   `gorm:"column:is_online;not null;default:0" json:"is_online"`                                             // 0-未上线 1-已上线
}
