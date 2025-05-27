package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.uber.org/zap"

	"kube-tide/internal/database"
)

func main() {
	var (
		action     = flag.String("action", "migrate", "Action to perform: migrate, rollback, version")
		version    = flag.String("version", "", "Target version for rollback")
		dbType     = flag.String("type", "postgres", "Database type: postgres or sqlite")
		dbHost     = flag.String("host", "localhost", "Database host")
		dbPort     = flag.Int("port", 5432, "Database port")
		dbUser     = flag.String("user", "postgres", "Database user")
		dbPassword = flag.String("password", "", "Database password")
		dbName     = flag.String("database", "kube_tide", "Database name")
		sqliteFile = flag.String("sqlite-file", "./data/kube_tide.db", "SQLite file path")
	)
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Create database configuration
	config := &database.DatabaseConfig{
		Type:            database.DatabaseType(*dbType),
		Host:            *dbHost,
		Port:            *dbPort,
		User:            *dbUser,
		Password:        *dbPassword,
		Database:        *dbName,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		SQLiteFilePath:  *sqliteFile,
	}

	// Create database service
	dbService, err := database.NewService(config, logger)
	if err != nil {
		logger.Fatal("Failed to create database service", zap.Error(err))
	}
	defer dbService.Close()

	ctx := context.Background()

	switch *action {
	case "migrate":
		logger.Info("Running database migrations...")
		if err := dbService.Initialize(ctx); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Migrations completed successfully")

	case "version":
		version, err := dbService.GetCurrentMigrationVersion(ctx)
		if err != nil {
			logger.Fatal("Failed to get current version", zap.Error(err))
		}
		fmt.Printf("Current migration version: %d\n", version)

	case "rollback":
		if *version == "" {
			logger.Fatal("Version is required for rollback")
		}
		targetVersion, err := strconv.Atoi(*version)
		if err != nil {
			logger.Fatal("Invalid version number", zap.Error(err))
		}
		logger.Info("Rolling back database...", zap.Int("target_version", targetVersion))
		if err := dbService.RollbackToVersion(ctx, targetVersion); err != nil {
			logger.Fatal("Failed to rollback database", zap.Error(err))
		}
		logger.Info("Rollback completed successfully")

	default:
		logger.Fatal("Unknown action", zap.String("action", *action))
	}
}
