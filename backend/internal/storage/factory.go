package storage

import (
	"fmt"
	"voicepaper/config"
)

// NewStorage 根据配置创建存储实例
// 支持 local 和 oss 两种类型
func NewStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage.Type {
	case "local":
		return NewLocalStorage(cfg), nil
	case "oss":
		return NewOSSStorage(cfg)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s，支持的类型: local, oss", cfg.Storage.Type)
	}
}

