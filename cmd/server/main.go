package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kube-tide/configs"
	"kube-tide/internal/api"
	"kube-tide/internal/core/k8s"
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

	log := logger.GetLogger()
	defer log.Sync() // Make sure the buffer log is written before the program exits

	log.Info("初始化服务...")
	log.Info("配置加载完成",
		logger.String("端口", config.Server.Port),
		logger.Bool("日志文件启用", config.Logging.FileConfig.Enabled),
		logger.Bool("日志滚动启用", config.Logging.RotateConfig.Enabled),
	)

	// init kubernetes client
	clientManager := k8s.NewClientManager()

	// create services
	nodePoolService := k8s.NewNodePoolService(clientManager)
	nodeService := k8s.NewNodeService(clientManager, nodePoolService)
	podService := k8s.NewPodService(clientManager)
	deploymentService := k8s.NewDeploymentService(clientManager)
	serviceManager := k8s.NewServiceManager(clientManager)

	// create API handlers
	nodeHandler := api.NewNodeHandler(nodeService)
	podHandler := api.NewPodHandler(podService)
	deploymentHandler := api.NewDeploymentHandler(deploymentService)
	nodePoolHandler := api.NewNodePoolHandler(nodePoolService)
	serviceHandler := api.NewServiceHandler(serviceManager)
	clusterHandler := api.NewClusterHandler(clientManager)
	healthHandler := api.NewHealthCheckHandler()
	podTerminalHandler := api.NewPodTerminalHandler(podService)

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
		log.Info("服务器启动",
			logger.String("监听地址", srv.Addr),
			logger.String("API文档", fmt.Sprintf("http://%s:%s/api", config.Server.Host, config.Server.Port)),
			logger.String("Web界面", fmt.Sprintf("http://%s:%s", config.Server.Host, config.Server.Port)),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("监听失败", logger.Error(err))
		}
	}()

	// Wait for interrupt signal to close the server gracefully
	quit := make(chan os.Signal, 1)
	// kill (no parameters) sends syscall.SIGTERM
	// kill -2 sends syscall.SIGINT
	// kill -9 is syscall.SIGKILL, but cannot be caught, so no need to add
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down the server...")

	// Set a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", logger.Error(err))
	}

	log.Info("Server has exited safely")
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
