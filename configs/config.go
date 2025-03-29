package configs

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// Set at build time, indicating the build mode
var BuildMode string

// Config application configuration structure
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// ServerConfig Server configuration
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// LoggingConfig Logging configuration
type LoggingConfig struct {
	Level        string          `mapstructure:"level"`
	FileConfig   LogFileConfig   `mapstructure:"file"`
	RotateConfig LogRotateConfig `mapstructure:"rotate"`
}

// LogFileConfig File logging configuration
type LogFileConfig struct {
	Enabled   bool   `mapstructure:"enabled"`    // 是否启用文件日志
	Path      string `mapstructure:"path"`       // 日志文件路径
	ErrorPath string `mapstructure:"error_path"` // 错误日志文件路径
}

// LogRotateConfig Logging rotation configuration
type LogRotateConfig struct {
	Enabled      bool   `mapstructure:"enabled"`       // 是否启用日志滚动
	MaxSize      int    `mapstructure:"max_size"`      // 每个日志文件的最大大小，单位MB
	MaxAge       int    `mapstructure:"max_age"`       // 保留旧日志文件的最大天数
	MaxBackups   int    `mapstructure:"max_backups"`   // 保留的旧日志文件的最大数量
	Compress     bool   `mapstructure:"compress"`      // 是否压缩旧日志文件
	LocalTime    bool   `mapstructure:"local_time"`    // 使用本地时间命名备份文件
	RotationTime string `mapstructure:"rotation_time"` // 轮转时间间隔（daily/hourly）
}

// LoadConfig loads the configuration from the config file
func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("logging.level", "info")

	// Set default values for file logging
	viper.SetDefault("logging.file.enabled", false)
	viper.SetDefault("logging.file.path", "./logs/app.log")
	viper.SetDefault("logging.file.error_path", "./logs/error.log")

	// Set default values for log rotation
	viper.SetDefault("logging.rotate.enabled", false)
	viper.SetDefault("logging.rotate.max_size", 100)
	viper.SetDefault("logging.rotate.max_age", 30)
	viper.SetDefault("logging.rotate.max_backups", 10)
	viper.SetDefault("logging.rotate.compress", true)
	viper.SetDefault("logging.rotate.local_time", true)
	viper.SetDefault("logging.rotate.rotation_time", "daily")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: unable to read config file: %v", err)
		log.Println("Using default configuration")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Cannot unmarshal config: %v", err)
	}

	return &config
}

// IsDevMode is a function to check if the application is running in development mode
func IsDevMode() bool {
	// First, check the environment variable
	if os.Getenv("K8S_PLATFORM_ENV") == "production" {
		return false
	}

	// Then, check the build-time flag
	if BuildMode == "production" {
		return false
	}

	// Default to development mode
	return true
}

// IsProductionMode is a function to check if the application is running in production mode
func IsProductionMode() bool {
	return !IsDevMode()
}
