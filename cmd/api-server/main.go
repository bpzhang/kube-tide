package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-tide/internal/api"
	"kube-tide/internal/core"
	"kube-tide/internal/database"
	"kube-tide/internal/database/migrations"
	"kube-tide/internal/repository"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

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
	db, err := database.NewDatabase(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	migrationService := migrations.NewMigrationService(db, logger)
	if err := migrationService.Migrate(context.Background()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully")

	// Initialize repositories
	repos := repository.NewRepositories(db, logger)

	// Initialize core service
	coreService := core.NewService(repos, db, logger)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		if err := db.Health(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"services": gin.H{
				"database": "healthy",
			},
		})
	})

	// Setup database API routes
	dbRouter := api.NewDBRouter(coreService, logger)
	dbRouter.SetupRoutes(router)

	// API documentation endpoint
	router.GET("/api/v1/db", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Kube-Tide Database API",
			"version": "v1",
			"endpoints": gin.H{
				"deployments": gin.H{
					"POST /api/v1/db/deployments":                                                 "Create deployment",
					"GET /api/v1/db/deployments/:id":                                              "Get deployment by ID",
					"PUT /api/v1/db/deployments/:id":                                              "Update deployment",
					"DELETE /api/v1/db/deployments/:id":                                           "Delete deployment",
					"GET /api/v1/db/clusters/:cluster_id/deployments":                             "List deployments by cluster",
					"DELETE /api/v1/db/clusters/:cluster_id/deployments":                          "Delete all deployments in cluster",
					"GET /api/v1/db/clusters/:cluster_id/deployments/count":                       "Count deployments in cluster",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments":       "List deployments by namespace",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments/:name": "Get deployment by name",
					"DELETE /api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments":    "Delete all deployments in namespace",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments/count": "Count deployments in namespace",
				},
				"services": gin.H{
					"POST /api/v1/db/services":                                                 "Create service",
					"GET /api/v1/db/services/:id":                                              "Get service by ID",
					"PUT /api/v1/db/services/:id":                                              "Update service",
					"DELETE /api/v1/db/services/:id":                                           "Delete service",
					"GET /api/v1/db/clusters/:cluster_id/services":                             "List services by cluster",
					"DELETE /api/v1/db/clusters/:cluster_id/services":                          "Delete all services in cluster",
					"GET /api/v1/db/clusters/:cluster_id/services/count":                       "Count services in cluster",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/services":       "List services by namespace",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/services/:name": "Get service by name",
					"DELETE /api/v1/db/clusters/:cluster_id/namespaces/:namespace/services":    "Delete all services in namespace",
					"GET /api/v1/db/clusters/:cluster_id/namespaces/:namespace/services/count": "Count services in namespace",
				},
			},
		})
	})

	// Start server
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting server", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
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
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
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
