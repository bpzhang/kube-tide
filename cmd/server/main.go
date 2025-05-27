package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"kube-tide/configs"
	"kube-tide/internal/api"
	"kube-tide/internal/core"
	"kube-tide/internal/core/k8s"
	"kube-tide/internal/database"
	"kube-tide/internal/database/migrations"
	"kube-tide/internal/repository"
	"kube-tide/internal/utils/logger"

	"go.uber.org/zap/zapcore"
)

func main() {
	// load config
	config := configs.LoadConfig()

	// logger level
	logLevel := getLogLevel(config.Logging.Level)

	// init logger
	logger.Init(
		logger.WithDevelopment(configs.IsDevMode()),
		logger.WithLevel(logLevel),
		logger.WithFileConfig(config.Logging.FileConfig),
		logger.WithRotateConfig(config.Logging.RotateConfig),
	)

	// 使用新的日志接口
	logger.Info("初始化服务...")
	logger.Info("配置加载完成",
		"端口", config.Server.Port,
		"日志文件启用", config.Logging.FileConfig.Enabled,
		"日志滚动启用", config.Logging.RotateConfig.Enabled,
	)

	// init kubernetes client
	clientManager := k8s.NewClientManager()

	// create services
	nodePoolService := k8s.NewNodePoolService(clientManager)
	nodeService := k8s.NewNodeService(clientManager, nodePoolService)
	podService := k8s.NewPodService(clientManager)
	deploymentService := k8s.NewDeploymentService(clientManager)
	serviceManager := k8s.NewServiceManager(clientManager)
	namespaceService := k8s.NewNamespaceService(clientManager)     // 初始化命名空间服务
	statefulSetService := k8s.NewStatefulSetService(clientManager) // 初始化StatefulSet服务

	// 初始化Pod指标服务，用于收集和缓存监控数据
	podMetricsService := k8s.NewPodMetricsService(clientManager)

	// 启动后台任务，每1分钟定期收集所有集群的Pod指标
	ctx, cancelCollect := context.WithCancel(context.Background())

	// 获取所有集群并为每个集群启动指标收集
	clusters := clientManager.ListClusters()
	if len(clusters) > 0 {
		for _, clusterName := range clusters {
			logger.Info("启动Pod指标收集", "集群", clusterName)
			podMetricsService.StartPeriodicMetricsCollection(ctx, clusterName, 1*time.Minute)
		}
	} else {
		logger.Warn("无法启动Pod指标收集", "错误", nil)
	}

	// 启动定期清理过期缓存的任务
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				logger.Debug("清理过期的Pod指标缓存")
				podMetricsService.CleanExpiredMetricsCache()
			}
		}
	}()

	// Initialize database (optional, for data persistence)
	var dbService *core.Service
	var dbRouter *api.DBRouter

	// Check if database should be enabled
	if getEnv("ENABLE_DATABASE", "false") == "true" {
		// Initialize database configuration
		dbConfig := &database.DatabaseConfig{
			Type:            database.DatabaseType(getEnv("DB_TYPE", "sqlite")),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "kube_tide"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			SQLiteFilePath:  getEnv("DB_SQLITE_PATH", "./data/kube_tide.db"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		}

		// Initialize database
		db, err := database.NewDatabase(dbConfig, logger.GetZapLogger())
		if err != nil {
			logger.Error("Failed to initialize database", "error", err.Error())
		} else {
			// Run migrations
			migrationService := migrations.NewMigrationService(db, logger.GetZapLogger())
			if err := migrationService.Migrate(context.Background()); err != nil {
				logger.Error("Failed to run migrations", "error", err.Error())
			} else {
				logger.Info("Database migrations completed successfully")

				// Initialize repositories and core service
				repos := repository.NewRepositories(db, logger.GetZapLogger())
				dbService = core.NewService(repos, db, logger.GetZapLogger())
				dbRouter = api.NewDBRouter(dbService, logger.GetZapLogger())
			}
		}
	}

	// create API handlers
	nodeHandler := api.NewNodeHandler(nodeService)
	podHandler := api.NewPodHandler(podService)

	// Create deployment handler with optional database service
	var deploymentHandler *api.DeploymentHandler
	if dbService != nil {
		deploymentHandler = api.NewDeploymentHandler(deploymentService, dbService.DeploymentService())
	} else {
		deploymentHandler = api.NewDeploymentHandler(deploymentService, nil)
	}

	nodePoolHandler := api.NewNodePoolHandler(nodePoolService)

	// Create service handler with optional database service
	var serviceHandler *api.ServiceHandler
	if dbService != nil {
		serviceHandler = api.NewServiceHandler(serviceManager, dbService.ServiceService())
	} else {
		serviceHandler = api.NewServiceHandler(serviceManager, nil)
	}

	clusterHandler := api.NewClusterHandler(clientManager)
	healthHandler := api.NewHealthCheckHandler()
	podTerminalHandler := api.NewPodTerminalHandler(podService)
	namespaceHandler := api.NewNamespaceHandler(namespaceService)       // 初始化命名空间处理器
	statefulSetHandler := api.NewStatefulSetHandler(statefulSetService) // 初始化StatefulSet处理器

	// Create an app instance and initialize the route
	app := &api.App{
		ClusterHandler:     clusterHandler,
		NodeHandler:        nodeHandler,
		PodHandler:         podHandler,
		ServiceHandler:     serviceHandler,
		DeploymentHandler:  deploymentHandler,
		NodePoolHandler:    nodePoolHandler,
		HealthHandler:      healthHandler,
		PodTerminalHandler: podTerminalHandler,
		NamespaceHandler:   namespaceHandler,   // 添加到App实例
		StatefulSetHandler: statefulSetHandler, // 添加StatefulSet处理器到App实例
		DBRouter:           dbRouter,           // 添加数据库路由器
	}

	// Initialize the router defined in router.go
	r := api.InitRouter(app)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Server.Port),
		Handler: r,
	}

	// Start the server in a separate goroutine
	go func() {
		logger.Info("服务器启动",
			"监听地址", srv.Addr,
			"API文档", fmt.Sprintf("http://%s:%s/api", config.Server.Host, config.Server.Port),
			"Web界面", fmt.Sprintf("http://%s:%s", config.Server.Host, config.Server.Port),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("监听失败", "error", err.Error())
		}
	}()

	// Wait for interrupt signal to close the server gracefully
	quit := make(chan os.Signal, 1)
	// kill (no parameters) sends syscall.SIGTERM
	// kill -2 sends syscall.SIGINT
	// kill -9 is syscall.SIGKILL, but cannot be caught, so no need to add
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down the server...")

	// 停止指标收集
	cancelCollect()
	logger.Info("已停止指标数据收集")

	// Set a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err.Error())
	}

	logger.Info("Server has exited safely")
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
