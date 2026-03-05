package model

import (
	"time"
)

// MediaResourceType 媒体资源类型
type MediaResourceType string

const (
	MediaResourceTypeAudio    MediaResourceType = "audio"
	MediaResourceTypeTimeline MediaResourceType = "timeline"
)

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeOSS   StorageType = "oss"
	StorageTypeS3    StorageType = "s3"
	StorageTypeCOS   StorageType = "cos"
)

// MediaResourceStatus 资源状态
type MediaResourceStatus string

const (
	MediaResourceStatusPending    MediaResourceStatus = "pending"
	MediaResourceStatusUploading  MediaResourceStatus = "uploading"
	MediaResourceStatusCompleted  MediaResourceStatus = "completed"
	MediaResourceStatusFailed     MediaResourceStatus = "failed"
)

// MediaResource 媒体资源表（统一管理音频和时间线文件）
type MediaResource struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ArticleID    uint              `gorm:"index;not null" json:"article_id"`
	ResourceType MediaResourceType `gorm:"size:20;not null" json:"resource_type"` // 'audio' | 'timeline'
	StorageType  StorageType       `gorm:"size:20;not null" json:"storage_type"`  // 'local' | 'oss' | 's3' | 'cos'

	// 存储路径/URL
	StoragePath string `gorm:"size:512;not null" json:"storage_path"` // 本地路径或 OSS key
	StorageURL  string `gorm:"size:512" json:"storage_url"`          // 完整访问 URL（OSS 使用）

	// 文件元信息
	FileName string `gorm:"size:255;not null" json:"file_name"`
	FileSize int64  `json:"file_size"`                    // 字节
	MimeType string `gorm:"size:100" json:"mime_type"`    // 'audio/mpeg', 'application/json'
	FileHash string `gorm:"size:64;index" json:"file_hash"` // 文件 MD5/SHA256，用于去重和校验

	// OSS 配置（如果使用对象存储）
	OSSBucket string `gorm:"size:100" json:"oss_bucket"`
	OSSRegion string `gorm:"size:50" json:"oss_region"`

	// 状态
	Status         MediaResourceStatus `gorm:"size:50;default:'pending'" json:"status"`
	UploadProgress int                `gorm:"default:0" json:"upload_progress"` // 上传进度 0-100

	// 关联
	Article        Article            `gorm:"foreignKey:ArticleID" json:"-"`
	TimelineSegments []TimelineSegment `gorm:"foreignKey:ResourceID" json:"-"`
}

// TimelineSegment 时间线片段
type TimelineSegment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	ArticleID  uint   `gorm:"index;not null" json:"article_id"`
	ResourceID uint   `gorm:"index;not null" json:"resource_id"` // 关联 media_resources.id

	// 时间信息（毫秒）
	TimeBegin int64 `gorm:"not null" json:"time_begin"`
	TimeEnd   int64 `gorm:"not null" json:"time_end"`

	// 文本信息
	Text      string `gorm:"type:text;not null" json:"text"`
	TextBegin int    `gorm:"not null" json:"text_begin"` // 在原文中的起始位置
	TextEnd   int    `gorm:"not null" json:"text_end"`    // 在原文中的结束位置

	// 排序
	SegmentOrder int `gorm:"not null" json:"segment_order"`

	// 关联
	Article  Article       `gorm:"foreignKey:ArticleID" json:"-"`
	Resource MediaResource `gorm:"foreignKey:ResourceID" json:"-"`
}

