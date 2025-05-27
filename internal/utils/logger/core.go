package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"kube-tide/configs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
func getWriter(path string, config lumberjackConfig) zapcore.WriteSyncer {
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

	// 解析压缩策略
	compress := false
	var compressDays int

	if strings.HasPrefix(config.Compression, "after_days:") {
		// 提取天数
		daysStr := strings.TrimPrefix(config.Compression, "after_days:")
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
			compressDays = days
		}
	} else if config.Compression == "immediate" {
		compress = true
	}

	// 配置日志滚动
	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSize, // 最大文件大小（MB）
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   compress, // 只有在immediate模式下才设置为true
		LocalTime:  config.LocalTime,
	}

	// 处理基于时间的轮转
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

				// 如果配置了延迟压缩，则处理旧日志文件的压缩
				if compressDays > 0 {
					go compressOldLogFiles(dir, nameWithoutExt, ext, compressDays)
				}
			}
		}()
	}

	// 如果只配置了延迟压缩但没有基于时间的轮转，也需要定期检查旧日志文件
	if compressDays > 0 && config.RotationTime == "" {
		go func() {
			// 每天检查一次旧日志
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()

			for range ticker.C {
				dir := filepath.Dir(path)
				base := filepath.Base(path)
				ext := filepath.Ext(base)
				nameWithoutExt := base[:len(base)-len(ext)]

				compressOldLogFiles(dir, nameWithoutExt, ext, compressDays)
			}
		}()
	}

	return zapcore.AddSync(logger)
}

// compressOldLogFiles 压缩超过指定天数的日志文件
func compressOldLogFiles(logDir string, namePrefix string, ext string, daysOld int) {
	// 确保目录存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return
	}

	// 获取当前时间
	now := time.Now()
	cutoffTime := now.AddDate(0, 0, -daysOld)

	// 遍历日志目录
	files, err := os.ReadDir(logDir)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to read log directory: %v", err)
		return
	}

	// 日志文件命名格式: namePrefix-YYYY-MM-DD.ext
	datePattern := fmt.Sprintf(`%s-(\d{4}-\d{2}-\d{2})%s`, namePrefix, ext)
	dateRegex, err := regexp.Compile(datePattern)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to compile regex pattern: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()

		// 跳过已压缩的文件
		if strings.HasSuffix(filename, ".gz") || strings.HasSuffix(filename, ".zip") {
			continue
		}

		// 检查是否匹配日志文件模式
		matches := dateRegex.FindStringSubmatch(filename)
		if len(matches) < 2 {
			continue
		}

		// 解析文件日期
		dateStr := matches[1]
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// 检查文件是否超过指定天数
		if fileDate.Before(cutoffTime) {
			filePath := filepath.Join(logDir, filename)
			compressLogFile(filePath)
		}
	}
}

// compressLogFile 压缩单个日志文件
func compressLogFile(filePath string) {
	globalSugaredLogger.Infof("Compressing log file: %s", filePath)

	// 打开源文件
	sourceFile, err := os.Open(filePath)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to open file for compression: %v", err)
		return
	}
	defer sourceFile.Close()

	// 创建压缩文件
	targetPath := filePath + ".gz"
	targetFile, err := os.Create(targetPath)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to create compressed file: %v", err)
		return
	}
	defer targetFile.Close()

	// 创建gzip写入器
	gzipWriter := gzip.NewWriter(targetFile)
	defer gzipWriter.Close()

	// 复制内容
	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to compress file: %v", err)
		// 如果压缩失败，删除不完整的压缩文件
		targetFile.Close()
		os.Remove(targetPath)
		return
	}

	// 确保所有数据都刷新到压缩文件
	gzipWriter.Flush()
	gzipWriter.Close()

	// 压缩成功后删除原始文件
	sourceFile.Close()
	err = os.Remove(filePath)
	if err != nil {
		globalSugaredLogger.Errorf("Failed to remove original file after compression: %v", err)
	}
}

// 获取zap Logger实例
func getZapLogger() *zap.Logger {
	if globalLogger == nil {
		Init()
	}
	return globalLogger
}

// GetZapLogger 公开的获取zap Logger实例的方法
func GetZapLogger() *zap.Logger {
	return getZapLogger()
}

// 获取SugaredLogger实例
func getZapSugaredLogger() *zap.SugaredLogger {
	if globalSugaredLogger == nil {
		Init()
	}
	return globalSugaredLogger
}

// GetZapSugaredLogger 公开的获取SugaredLogger实例的方法
func GetZapSugaredLogger() *zap.SugaredLogger {
	return getZapSugaredLogger()
}

// 将日志记录接口转换为字段
func toZapFields(args ...interface{}) []zapcore.Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]zapcore.Field, 0, len(args)/2)

	for i := 0; i < len(args); i += 2 {
		// 确保有键值对
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				key = fmt.Sprintf("%v", args[i])
			}
			fields = append(fields, zap.Any(key, args[i+1]))
		}
	}

	return fields
}

// 类型别名，隐藏内部实现细节
type lumberjackConfig = configs.LogRotateConfig
