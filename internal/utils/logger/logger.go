package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"kube-tide/configs"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// 全局logger实例
	globalLogger *zap.Logger
	// 全局sugaredLogger实例
	globalSugaredLogger *zap.SugaredLogger
	// 确保只初始化一次的互斥锁
	once sync.Once
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
		Compress:     true,
		LocalTime:    true,
		RotationTime: "daily",
	},
}

// 初始化日志
func Init(opts ...func(*Options)) {
	once.Do(func() {
		// 应用配置选项
		options := defaultOptions
		for _, opt := range opts {
			opt(&options)
		}

		// 创建编码器配置
		var encoderConfig zapcore.EncoderConfig
		if options.Development {
			encoderConfig = zap.NewDevelopmentEncoderConfig()
		} else {
			encoderConfig = zap.NewProductionEncoderConfig()
		}
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		// 设置日志输出
		var cores []zapcore.Core

		// 控制台输出
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= options.Level
		})
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), consoleLevel)
		cores = append(cores, consoleCore)

		// 文件输出配置（如果启用）
		if options.FileConfig.Enabled {
			// 确保日志目录存在
			ensureLogDir(options.FileConfig.Path)
			ensureLogDir(options.FileConfig.ErrorPath)

			// 创建日志滚动配置
			regularSyncer := zapcore.AddSync(getWriter(options.FileConfig.Path, options.RotateConfig))
			errorSyncer := zapcore.AddSync(getWriter(options.FileConfig.ErrorPath, options.RotateConfig))

			// 文件编码器（JSON格式，更适合日志处理）
			fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

			// 日志级别过滤器
			regularLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= options.Level && lvl < zapcore.ErrorLevel
			})
			errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			})

			// 创建核心
			regularCore := zapcore.NewCore(fileEncoder, regularSyncer, regularLevel)
			errorCore := zapcore.NewCore(fileEncoder, errorSyncer, errorLevel)

			cores = append(cores, regularCore, errorCore)
		}

		// 合并所有核心
		core := zapcore.NewTee(cores...)

		// 创建logger
		globalLogger = zap.New(
			core,
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)

		// 创建SugaredLogger
		globalSugaredLogger = globalLogger.Sugar()
	})
}

// 确保日志目录存在
func ensureLogDir(path string) {
	dir := path[:strings.LastIndex(path, "/")]
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// 创建日志写入器，支持滚动
func getWriter(path string, config configs.LogRotateConfig) zapcore.WriteSyncer {
	if !config.Enabled {
		// 未启用滚动时，返回标准文件写入器
		file, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		return zapcore.AddSync(file)
	}

	// 如果启用了按时间轮转，则使用带时间戳的文件名
	var filename string
	var ext string
	if config.RotationTime != "" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		ext = filepath.Ext(base)
		nameWithoutExt := base[:len(base)-len(ext)]

		// 使用当前日期创建文件名
		now := time.Now()
		dateStr := now.Format("2006-01-02")
		filename = filepath.Join(dir, fmt.Sprintf("%s-%s%s", nameWithoutExt, dateStr, ext))
	} else {
		filename = path
	}

	// 配置日志滚动
	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSize, // 最大文件大小（MB）
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  config.LocalTime,
	}

	// 如果配置了基于时间的轮转
	if config.RotationTime != "" {
		// 启动一个goroutine来处理基于时间的日志轮转
		go func() {
			for {
				now := time.Now()
				var next time.Time

				// 根据配置的轮转时间计算下一次轮转时间
				switch strings.ToLower(config.RotationTime) {
				case "hourly":
					next = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
				case "daily":
					next = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
				default:
					// 默认每天轮转
					next = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
				}

				duration := next.Sub(now)
				timer := time.NewTimer(duration)
				<-timer.C

				// 生成新的日志文件名
				dir := filepath.Dir(path)
				base := filepath.Base(path)
				ext := filepath.Ext(base)
				nameWithoutExt := base[:len(base)-len(ext)]

				newDate := time.Now().Format("2006-01-02")
				newFilename := filepath.Join(dir, fmt.Sprintf("%s-%s%s", nameWithoutExt, newDate, ext))

				// 关闭当前日志文件
				logger.Close()

				// 更新文件名并重新打开
				logger.Filename = newFilename
			}
		}()
	}

	return zapcore.AddSync(logger)
}

// 获取zap Logger实例
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		Init()
	}
	return globalLogger
}

// 获取SugaredLogger实例，提供更方便的接口
func GetSugaredLogger() *zap.SugaredLogger {
	if globalSugaredLogger == nil {
		Init()
	}
	return globalSugaredLogger
}

// 带有上下文的logger
func WithContext(fields ...zapcore.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

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

// Debug 全局Debug日志
func Debug(msg string, fields ...zapcore.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 全局Info日志
func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 全局Warn日志
func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 全局Error日志
func Err(msg string, fields ...zapcore.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 全局Fatal日志
func Fatal(msg string, fields ...zapcore.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debugf 全局Debug格式化日志
func Debugf(template string, args ...interface{}) {
	GetSugaredLogger().Debugf(template, args...)
}

// Infof 全局Info格式化日志
func Infof(template string, args ...interface{}) {
	GetSugaredLogger().Infof(template, args...)
}

// Warnf 全局Warn格式化日志
func Warnf(template string, args ...interface{}) {
	GetSugaredLogger().Warnf(template, args...)
}

// Errorf 全局Error格式化日志
func Errorf(template string, args ...interface{}) {
	GetSugaredLogger().Errorf(template, args...)
}

// Fatalf 全局Fatal格式化日志
func Fatalf(template string, args ...interface{}) {
	GetSugaredLogger().Fatalf(template, args...)
}

// 以下是字段构造函数，直接封装 zap 的同名函数，减少对 zap 包的直接依赖

// String 创建字符串字段
func String(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

// Int 创建整数字段
func Int(key string, val int) zapcore.Field {
	return zap.Int(key, val)
}

// Int64 创建int64字段
func Int64(key string, val int64) zapcore.Field {
	return zap.Int64(key, val)
}

// Float64 创建float64字段
func Float64(key string, val float64) zapcore.Field {
	return zap.Float64(key, val)
}

// Bool 创建布尔字段
func Bool(key string, val bool) zapcore.Field {
	return zap.Bool(key, val)
}

// Error 创建错误字段
func Error(err error) zapcore.Field {
	return zap.Error(err)
}

// Duration 创建时间段字段
func Duration(key string, val time.Duration) zapcore.Field {
	return zap.Duration(key, val)
}

// Time 创建时间字段
func Time(key string, val time.Time) zapcore.Field {
	return zap.Time(key, val)
}

// Object 创建对象字段
func Object(key string, val interface{}) zapcore.Field {
	return zap.Any(key, val)
}

// Any 创建任意类型字段
func Any(key string, val interface{}) zapcore.Field {
	return zap.Any(key, val)
}

// Reflect 创建反射字段
func Reflect(key string, val interface{}) zapcore.Field {
	return zap.Reflect(key, val)
}
