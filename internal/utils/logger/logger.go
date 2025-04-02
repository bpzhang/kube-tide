package logger

import (
	"context"
	"time"
)

// Logger 提供一个简化的日志接口
type Logger interface {
	// 设置上下文
	WithContext(ctx context.Context) Logger

	// 基本日志方法 - 不暴露zapcore.Field
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	// 格式化日志方法
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

// 默认logger实现
type defaultLogger struct {
	ctx context.Context
}

// NewLogger 创建一个新的logger实例
func NewLogger() Logger {
	return &defaultLogger{}
}

// WithContext 设置上下文
func (l *defaultLogger) WithContext(ctx context.Context) Logger {
	return &defaultLogger{ctx: ctx}
}

// Debug 记录Debug级别日志
func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	getZapLogger().Debug(msg, toZapFields(args...)...)
}

// Info 记录Info级别日志
func (l *defaultLogger) Info(msg string, args ...interface{}) {
	getZapLogger().Info(msg, toZapFields(args...)...)
}

// Warn 记录Warn级别日志
func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	getZapLogger().Warn(msg, toZapFields(args...)...)
}

// Error 记录Error级别日志
func (l *defaultLogger) Error(msg string, args ...interface{}) {
	getZapLogger().Error(msg, toZapFields(args...)...)
}

// Fatal 记录Fatal级别日志
func (l *defaultLogger) Fatal(msg string, args ...interface{}) {
	getZapLogger().Fatal(msg, toZapFields(args...)...)
}

// Debugf 记录Debug级别格式化日志
func (l *defaultLogger) Debugf(template string, args ...interface{}) {
	getZapSugaredLogger().Debugf(template, args...)
}

// Infof 记录Info级别格式化日志
func (l *defaultLogger) Infof(template string, args ...interface{}) {
	getZapSugaredLogger().Infof(template, args...)
}

// Warnf 记录Warn级别格式化日志
func (l *defaultLogger) Warnf(template string, args ...interface{}) {
	getZapSugaredLogger().Warnf(template, args...)
}

// Errorf 记录Error级别格式化日志
func (l *defaultLogger) Errorf(template string, args ...interface{}) {
	getZapSugaredLogger().Errorf(template, args...)
}

// Fatalf 记录Fatal级别格式化日志
func (l *defaultLogger) Fatalf(template string, args ...interface{}) {
	getZapSugaredLogger().Fatalf(template, args...)
}

// 全局默认logger实例
var defaultLoggerInstance = NewLogger()

// 全局日志方法，不再暴露zapcore.Field类型

// Debug 全局Debug日志
func Debug(msg string, args ...interface{}) {
	defaultLoggerInstance.Debug(msg, args...)
}

// Info 全局Info日志
func Info(msg string, args ...interface{}) {
	defaultLoggerInstance.Info(msg, args...)
}

// Warn 全局Warn日志
func Warn(msg string, args ...interface{}) {
	defaultLoggerInstance.Warn(msg, args...)
}

// Error 全局Error日志
func Err(msg string, args ...interface{}) {
	defaultLoggerInstance.Error(msg, args...)
}

// Fatal 全局Fatal日志
func Fatal(msg string, args ...interface{}) {
	defaultLoggerInstance.Fatal(msg, args...)
}

// Debugf 全局Debug格式化日志
func Debugf(template string, args ...interface{}) {
	defaultLoggerInstance.Debugf(template, args...)
}

// Infof 全局Info格式化日志
func Infof(template string, args ...interface{}) {
	defaultLoggerInstance.Infof(template, args...)
}

// Warnf 全局Warn格式化日志
func Warnf(template string, args ...interface{}) {
	defaultLoggerInstance.Warnf(template, args...)
}

// Errorf 全局Error格式化日志
func Errorf(template string, args ...interface{}) {
	defaultLoggerInstance.Errorf(template, args...)
}

// Fatalf 全局Fatal格式化日志
func Fatalf(template string, args ...interface{}) {
	defaultLoggerInstance.Fatalf(template, args...)
}

// 以下是用于创建AOP风格日志切面的辅助方法

// LogOperation 记录操作的开始和结束，以及可能的错误
func LogOperation(operationName string, fn func() error) error {
	Infof("开始操作: %s", operationName)
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	if err != nil {
		Errorf("操作失败: %s, 耗时: %v, 错误: %v", operationName, duration, err)
	} else {
		Infof("操作成功: %s, 耗时: %v", operationName, duration)
	}

	return err
}

// LogFunc 为任意函数添加日志记录装饰
func LogFunc(funcName string, fn func() (interface{}, error)) (interface{}, error) {
	Debugf("调用函数: %s", funcName)
	start := time.Now()

	result, err := fn()

	duration := time.Since(start)
	if err != nil {
		Errorf("函数调用失败: %s, 耗时: %v, 错误: %v", funcName, duration, err)
	} else {
		Debugf("函数调用成功: %s, 耗时: %v", funcName, duration)
	}

	return result, err
}

// LogWithContext 为带上下文的操作添加日志记录
func LogWithContext(ctx context.Context, operationName string, fn func(ctx context.Context) error) error {
	logger := NewLogger().WithContext(ctx)
	logger.Infof("开始操作: %s", operationName)
	start := time.Now()

	err := fn(ctx)

	duration := time.Since(start)
	if err != nil {
		logger.Errorf("操作失败: %s, 耗时: %v, 错误: %v", operationName, duration, err)
	} else {
		logger.Infof("操作成功: %s, 耗时: %v", operationName, duration)
	}

	return err
}

// K8sOperationLogger K8s专用日志记录器
type K8sOperationLogger struct {
	ClusterName  string
	Namespace    string
	ResourceKind string
}

// NewK8sLogger 创建K8s操作日志记录器
func NewK8sLogger(clusterName, namespace, resourceKind string) *K8sOperationLogger {
	return &K8sOperationLogger{
		ClusterName:  clusterName,
		Namespace:    namespace,
		ResourceKind: resourceKind,
	}
}

// LogOperation 记录K8s操作
func (kl *K8sOperationLogger) LogOperation(operation string, fn func() (interface{}, error)) (interface{}, error) {
	clusterInfo := kl.ClusterName
	if kl.Namespace != "" {
		clusterInfo += ":" + kl.Namespace
	}

	Info("K8s操作开始",
		"cluster", kl.ClusterName,
		"namespace", kl.Namespace,
		"resource", kl.ResourceKind,
		"operation", operation,
	)

	start := time.Now()
	result, err := fn()
	duration := time.Since(start)

	if err != nil {
		Err("K8s操作失败",
			"cluster", kl.ClusterName,
			"namespace", kl.Namespace,
			"resource", kl.ResourceKind,
			"operation", operation,
			"duration", duration.String(),
			"error", err.Error(),
		)
	} else {
		Info("K8s操作成功",
			"cluster", kl.ClusterName,
			"namespace", kl.Namespace,
			"resource", kl.ResourceKind,
			"operation", operation,
			"duration", duration.String(),
		)
	}

	return result, err
}

// LogRequestOperation 用于记录HTTP请求处理过程
func LogRequestOperation(requestPath, method, operationName string, fn func() error) error {
	Info("处理请求",
		"path", requestPath,
		"method", method,
		"operation", operationName,
	)

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	if err != nil {
		Err("请求处理失败",
			"path", requestPath,
			"method", method,
			"operation", operationName,
			"duration", duration.String(),
			"error", err.Error(),
		)
	} else {
		Info("请求处理成功",
			"path", requestPath,
			"method", method,
			"operation", operationName,
			"duration", duration.String(),
		)
	}

	return err
}
