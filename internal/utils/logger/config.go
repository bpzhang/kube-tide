package logger

import (
	"kube-tide/configs"

	"go.uber.org/zap/zapcore"
)

// 配置选项
type Options struct {
	// 开发模式 - 开发模式下使用彩色日志，否则使用JSON格式
	Development bool
	// 日志级别
	Level zapcore.Level
	// 日志输出路径，默认为标准输出
	OutputPaths []string
	// 错误日志输出路径
	ErrorOutputPaths []string
	// 文件日志配置
	FileConfig configs.LogFileConfig
	// 日志滚动配置
	RotateConfig configs.LogRotateConfig
}

// 默认配置
var defaultOptions = Options{
	Development:      true,
	Level:            zapcore.InfoLevel,
	OutputPaths:      []string{"stdout"},
	ErrorOutputPaths: []string{"stderr"},
	FileConfig: configs.LogFileConfig{
		Enabled:   false,
		Path:      "./logs/app.log",
		ErrorPath: "./logs/error.log",
	},
	RotateConfig: configs.LogRotateConfig{
		Enabled:      false,
		MaxSize:      100,
		MaxAge:       30,
		MaxBackups:   10,
		Compression:  "after_days:7",
		LocalTime:    true,
		RotationTime: "daily",
	},
}

// 配置选项修改器

// 设置开发模式
func WithDevelopment(dev bool) func(*Options) {
	return func(o *Options) {
		o.Development = dev
	}
}

// 设置日志级别
func WithLevel(level zapcore.Level) func(*Options) {
	return func(o *Options) {
		o.Level = level
	}
}

// 设置输出路径
func WithOutputPaths(paths ...string) func(*Options) {
	return func(o *Options) {
		o.OutputPaths = paths
	}
}

// 设置错误输出路径
func WithErrorOutputPaths(paths ...string) func(*Options) {
	return func(o *Options) {
		o.ErrorOutputPaths = paths
	}
}

// 设置文件日志配置
func WithFileConfig(config configs.LogFileConfig) func(*Options) {
	return func(o *Options) {
		o.FileConfig = config
	}
}

// 设置日志滚动配置
func WithRotateConfig(config configs.LogRotateConfig) func(*Options) {
	return func(o *Options) {
		o.RotateConfig = config
	}
}
