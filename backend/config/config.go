package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 应用配置结构
type Config struct {
	MiniMax  MiniMaxConfig  `yaml:"minimax"`
	TTS      TTSConfig      `yaml:"tts"`
	Storage  StorageConfig  `yaml:"storage"`
	Network  NetworkConfig  `yaml:"network"`
	Logging  LoggingConfig  `yaml:"logging"`
	Database DatabaseConfig `yaml:"database"`
	Service  ServiceConfig  `yaml:"service"`
	Auth     AuthConfig     `yaml:"auth"`
	Redis    RedisConfig    `yaml:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MiniMaxConfig MiniMax API 配置
type MiniMaxConfig struct {
	APIKey      string `yaml:"api_key"`
	BaseURL     string `yaml:"base_url"`
	QueryURL    string `yaml:"query_url"`
	RetrieveURL string `yaml:"retrieve_url"`
}

// TTSConfig 语音合成配置
type TTSConfig struct {
	Model           string  `yaml:"model"`
	VoiceID         string  `yaml:"voice_id"`
	Speed           float64 `yaml:"speed"`
	Volume          float64 `yaml:"volume"`
	Pitch           int     `yaml:"pitch"`
	AudioSampleRate int     `yaml:"audio_sample_rate"`
	Bitrate         int     `yaml:"bitrate"`
	Format          string  `yaml:"format"`
	Channel         int     `yaml:"channel"`
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	Type      string    `yaml:"type"`       // 'local' | 'oss'
	OutputDir string    `yaml:"output_dir"` // 本地存储目录
	TempDir   string    `yaml:"temp_dir"`   // 临时目录
	OSS       OSSConfig `yaml:"oss"`        // OSS配置
}

// OSSConfig 阿里云OSS配置
type OSSConfig struct {
	Endpoint        string `yaml:"endpoint"`          // OSS endpoint，如 oss-cn-chengdu.aliyuncs.com
	AccessKeyID     string `yaml:"access_key_id"`     // AccessKey ID
	AccessKeySecret string `yaml:"access_key_secret"` // AccessKey Secret
	Bucket          string `yaml:"bucket"`            // Bucket名称
	Region          string `yaml:"region"`            // 区域，如 cn-chengdu
	BaseURL         string `yaml:"base_url"`          // 自定义域名（可选），如 https://cdn.example.com
	UseHTTPS        bool   `yaml:"use_https"`         // 是否使用HTTPS，默认true
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Timeout    int `yaml:"timeout"`
	RetryCount int `yaml:"retry_count"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Charset         string `yaml:"charset"`
	ParseTime       bool   `yaml:"parse_time"`
	Loc             string `yaml:"loc"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // 秒
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Port        string `yaml:"port"`
	Debug       bool   `yaml:"debug"`
	FrontendURL string `yaml:"frontend_url"` // 前端地址（用于OAuth重定向）
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWT     JWTConfig     `yaml:"jwt"`
	SMS     SMSConfig     `yaml:"sms"`
	Email   EmailConfig   `yaml:"email"`
	OAuth   OAuthConfig   `yaml:"oauth"`
	Session SessionConfig `yaml:"session"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `yaml:"secret"`     // JWT密钥
	Expiration int    `yaml:"expiration"` // 过期时间（小时）
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider string     `yaml:"provider"` // "aliyun" | "tencent"
	Aliyun   AliyunSMS  `yaml:"aliyun"`
	Tencent  TencentSMS `yaml:"tencent"`
}

// AliyunSMS 阿里云短信配置
type AliyunSMS struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	SignName        string `yaml:"sign_name"`     // 短信签名
	TemplateCode    string `yaml:"template_code"` // 验证码模板CODE
	Region          string `yaml:"region"`        // 区域，如 cn-hangzhou
}

// TencentSMS 腾讯云短信配置
type TencentSMS struct {
	SecretID   string `yaml:"secret_id"`
	SecretKey  string `yaml:"secret_key"`
	AppID      string `yaml:"app_id"`
	SignName   string `yaml:"sign_name"`   // 短信签名
	TemplateID string `yaml:"template_id"` // 模板ID
}

// EmailConfig 邮箱配置
type EmailConfig struct {
	SMTPHost     string `yaml:"smtp_host"`     // SMTP服务器地址
	SMTPPort     int    `yaml:"smtp_port"`     // SMTP端口
	SMTPUser     string `yaml:"smtp_user"`     // SMTP用户名（邮箱地址）
	SMTPPassword string `yaml:"smtp_password"` // SMTP密码（授权码）
	FromName     string `yaml:"from_name"`     // 发件人名称
	FromEmail    string `yaml:"from_email"`    // 发件人邮箱
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	GitHub GitHubOAuth `yaml:"github"`
	Google GoogleOAuth `yaml:"google"`
	WeChat WeChatOAuth `yaml:"wechat"`
}

// GitHubOAuth GitHub OAuth配置
type GitHubOAuth struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// GoogleOAuth Google OAuth配置
type GoogleOAuth struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// WeChatOAuth 微信小程序 OAuth配置
type WeChatOAuth struct {
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	Secret     string `yaml:"secret"`     // 会话密钥
	Expiration int    `yaml:"expiration"` // 过期时间（小时）
}

var AppConfig *Config

// LoadConfig 加载配置文件
// 支持从多个位置查找配置文件：
// 1. 环境变量 CONFIG_PATH 指定的路径（优先级最高）
// 2. 环境变量 ENV=pro 时，自动加载 config_pro.yaml
// 3. 根据域名自动判断：如果 frontend_url 包含 voicepaper.online，使用 config_pro.yaml
// 4. 当前目录下的 etc/config.yaml 或 etc/config_pro.yaml（优先）
// 5. 当前目录下的 config/config.yaml 或 config/config_pro.yaml
// 6. 默认使用 etc/config.yaml
func LoadConfig(configPath string) (*Config, error) {
	// 如果未指定路径，尝试自动查找
	if configPath == "" {
		// 优先级1: 环境变量 CONFIG_PATH
		if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
			configPath = envPath
			log.Printf("ℹ️  使用环境变量 CONFIG_PATH: %s", configPath)
		} else {
			// 优先级2: 检查环境变量 ENV
			env := os.Getenv("ENV")
			if env == "pro" || env == "prod" || env == "production" {
				// 生产环境，尝试加载 config_pro.yaml
				configPath = findConfigFile("config_pro.yaml")
				if configPath != "" {
					log.Printf("ℹ️  检测到生产环境 (ENV=%s)，使用配置文件: %s", env, configPath)
				}
			}

			// 如果还没找到，尝试加载默认的 config.yaml
			if configPath == "" {
				configPath = findConfigFile("config.yaml")
			}

			// 如果还是没找到，使用默认路径
			if configPath == "" {
				configPath = "etc/config.yaml"
			}
		}
	}

	// 第一次加载配置（可能是测试环境配置）
	firstConfig, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	// 根据 frontend_url 判断是否需要切换配置文件
	// 如果 frontend_url 包含 voicepaper.online，说明是生产环境，应该使用 config_pro.yaml
	if firstConfig.Service.FrontendURL != "" {
		if contains(firstConfig.Service.FrontendURL, "voicepaper.online") {
			// 生产环境，尝试加载 config_pro.yaml
			prodConfigPath := findConfigFile("config_pro.yaml")
			if prodConfigPath != "" && prodConfigPath != configPath {
				log.Printf("ℹ️  检测到生产环境域名 (frontend_url=%s)，切换到配置文件: %s", firstConfig.Service.FrontendURL, prodConfigPath)
				// 重新加载生产环境配置
				return loadConfigFile(prodConfigPath)
			}
		} else if contains(firstConfig.Service.FrontendURL, "localhost") {
			// 测试环境，确保使用 config.yaml
			testConfigPath := findConfigFile("config.yaml")
			if testConfigPath != "" && testConfigPath != configPath {
				log.Printf("ℹ️  检测到测试环境域名 (frontend_url=%s)，使用配置文件: %s", firstConfig.Service.FrontendURL, testConfigPath)
				return loadConfigFile(testConfigPath)
			}
		}
	}

	return firstConfig, nil
}

// loadConfigFile 加载指定的配置文件
func loadConfigFile(configPath string) (*Config, error) {

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	// 从环境变量覆盖敏感信息（环境变量优先级更高）
	overrideFromEnv(&config)

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	AppConfig = &config
	log.Printf("✅ 配置文件加载成功: %s", configPath)
	return &config, nil
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// findConfigFile 查找配置文件
// 按以下顺序查找：
// 1. etc/文件名（优先，新位置）
// 2. config/文件名（旧位置，兼容）
// 3. ../backend/etc/文件名
// 4. ../backend/config/文件名
// 5. 返回空字符串（表示未找到）
func findConfigFile(filename string) string {
	// 优先查找 etc/ 目录（新位置）
	paths := []string{
		filepath.Join("etc", filename),               // 当前目录下的 etc/
		filepath.Join("config", filename),            // 当前目录下的 config/（兼容旧位置）
		filepath.Join("../backend/etc", filename),    // 项目根目录下的 backend/etc/
		filepath.Join("../backend/config", filename), // 项目根目录下的 backend/config/
		filepath.Join("backend/etc", filename),       // backend/etc/
		filepath.Join("backend/config", filename),    // backend/config/
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// overrideFromEnv 从环境变量覆盖敏感配置
// 环境变量的优先级高于配置文件，这样可以更安全地管理敏感信息
func overrideFromEnv(c *Config) {
	// MiniMax API Key
	if apiKey := os.Getenv("MINIMAX_API_KEY"); apiKey != "" {
		c.MiniMax.APIKey = apiKey
		log.Println("ℹ️  使用环境变量 MINIMAX_API_KEY")
	}

	// 数据库配置
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		// 这里可以添加端口解析逻辑，暂时跳过
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		c.Database.Database = dbName
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		c.Database.Username = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		c.Database.Password = dbPassword
		log.Println("ℹ️  使用环境变量 DB_PASSWORD")
	}

	// OSS配置
	if ossEndpoint := os.Getenv("OSS_ENDPOINT"); ossEndpoint != "" {
		c.Storage.OSS.Endpoint = ossEndpoint
	}
	if ossAccessKeyID := os.Getenv("OSS_ACCESS_KEY_ID"); ossAccessKeyID != "" {
		c.Storage.OSS.AccessKeyID = ossAccessKeyID
		log.Println("ℹ️  使用环境变量 OSS_ACCESS_KEY_ID")
	}
	if ossAccessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET"); ossAccessKeySecret != "" {
		c.Storage.OSS.AccessKeySecret = ossAccessKeySecret
		log.Println("ℹ️  使用环境变量 OSS_ACCESS_KEY_SECRET")
	}
	if ossBucket := os.Getenv("OSS_BUCKET"); ossBucket != "" {
		c.Storage.OSS.Bucket = ossBucket
	}
	if ossRegion := os.Getenv("OSS_REGION"); ossRegion != "" {
		c.Storage.OSS.Region = ossRegion
	}

	// 认证配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.Auth.JWT.Secret = jwtSecret
		log.Println("ℹ️  使用环境变量 JWT_SECRET")
	}
	if aliyunSMSKeyID := os.Getenv("ALIYUN_SMS_ACCESS_KEY_ID"); aliyunSMSKeyID != "" {
		c.Auth.SMS.Aliyun.AccessKeyID = aliyunSMSKeyID
	}
	if aliyunSMSKeySecret := os.Getenv("ALIYUN_SMS_ACCESS_KEY_SECRET"); aliyunSMSKeySecret != "" {
		c.Auth.SMS.Aliyun.AccessKeySecret = aliyunSMSKeySecret
		log.Println("ℹ️  使用环境变量 ALIYUN_SMS_ACCESS_KEY_SECRET")
	}
	if emailSMTPPassword := os.Getenv("EMAIL_SMTP_PASSWORD"); emailSMTPPassword != "" {
		c.Auth.Email.SMTPPassword = emailSMTPPassword
		log.Println("ℹ️  使用环境变量 EMAIL_SMTP_PASSWORD")
	}
	if githubClientID := os.Getenv("GITHUB_CLIENT_ID"); githubClientID != "" {
		c.Auth.OAuth.GitHub.ClientID = githubClientID
	}
	if githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET"); githubClientSecret != "" {
		c.Auth.OAuth.GitHub.ClientSecret = githubClientSecret
		log.Println("ℹ️  使用环境变量 GITHUB_CLIENT_SECRET")
	}
	if googleClientID := os.Getenv("GOOGLE_CLIENT_ID"); googleClientID != "" {
		c.Auth.OAuth.Google.ClientID = googleClientID
	}
	if googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET"); googleClientSecret != "" {
		c.Auth.OAuth.Google.ClientSecret = googleClientSecret
		log.Println("ℹ️  使用环境变量 GOOGLE_CLIENT_SECRET")
	}

	if wechatAppID := os.Getenv("WECHAT_APP_ID"); wechatAppID != "" {
		c.Auth.OAuth.WeChat.AppID = wechatAppID
	}
	if wechatAppSecret := os.Getenv("WECHAT_APP_SECRET"); wechatAppSecret != "" {
		c.Auth.OAuth.WeChat.AppSecret = wechatAppSecret
		log.Println("ℹ️  使用环境变量 WECHAT_APP_SECRET")
	}

	// Redis配置
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		c.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		// 端口解析逻辑略
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		c.Redis.Password = redisPassword
		log.Println("ℹ️  使用环境变量 REDIS_PASSWORD")
	}
	// REDIS_DB 默认为0，通常不需要通过环境变量覆盖
}

// setDefaults 设置默认值
func setDefaults(c *Config) {
	if c.Storage.Type == "" {
		c.Storage.Type = "local" // 默认使用本地存储
	}
	if c.Storage.OutputDir == "" {
		c.Storage.OutputDir = "./data"
	}
	if c.Storage.TempDir == "" {
		c.Storage.TempDir = "./temp"
	}
	if c.Storage.OSS.UseHTTPS {
		c.Storage.OSS.UseHTTPS = true // 默认使用HTTPS
	}
	if c.Network.Timeout == 0 {
		c.Network.Timeout = 30
	}
	if c.Network.RetryCount == 0 {
		c.Network.RetryCount = 3
	}
	if c.Service.Port == "" {
		c.Service.Port = ":8080"
	} else if c.Service.Port[0] != ':' {
		c.Service.Port = ":" + c.Service.Port
	}
	// 设置默认前端URL（本地开发）
	if c.Service.FrontendURL == "" {
		c.Service.FrontendURL = "http://localhost:5173"
	}
	// 认证配置默认值
	if c.Auth.JWT.Secret == "" {
		c.Auth.JWT.Secret = "voicepaper-secret-key-change-in-production" // 生产环境必须修改
	}
	if c.Auth.JWT.Expiration == 0 {
		c.Auth.JWT.Expiration = 168 // 7天
	}
	if c.Auth.Session.Secret == "" {
		c.Auth.Session.Secret = "voicepaper-session-secret-change-in-production"
	}
	if c.Auth.Session.Expiration == 0 {
		c.Auth.Session.Expiration = 168 // 7天
	}
	if c.Auth.Email.SMTPPort == 0 {
		c.Auth.Email.SMTPPort = 587
	}
}

// validateConfig 验证配置
func validateConfig(c *Config) error {
	if c.MiniMax.APIKey == "" {
		return fmt.Errorf("MiniMax API Key 不能为空")
	}
	if c.MiniMax.BaseURL == "" {
		return fmt.Errorf("MiniMax BaseURL 不能为空")
	}
	if c.TTS.Model == "" {
		return fmt.Errorf("TTS Model 不能为空")
	}
	if c.TTS.VoiceID == "" {
		return fmt.Errorf("TTS VoiceID 不能为空")
	}
	// 认证配置验证（可选，如果启用认证功能）
	// 注意：认证功能是可选的，所以这里不强制验证
	return nil
}

// GetConfig 获取全局配置（如果已加载）
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("❌ 配置未加载，请先调用 LoadConfig()")
	}
	return AppConfig
}

// EnsureDirectories 确保配置中的目录存在
func EnsureDirectories(c *Config) error {
	dirs := []string{
		c.Storage.OutputDir,
		filepath.Join(c.Storage.OutputDir, "audio"),
		c.Storage.TempDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}

	return nil
}
