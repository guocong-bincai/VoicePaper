package repository

import (
	"voicepaper/internal/model"
)

type AppConfigRepository struct{}

func NewAppConfigRepository() *AppConfigRepository {
	return &AppConfigRepository{}
}

// GetConfigByKey 获取指定Key的配置
func (r *AppConfigRepository) GetConfigByKey(key string) (*model.AppConfig, error) {
	var config model.AppConfig
	err := DB.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
