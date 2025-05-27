package database

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"kube-tide/internal/database/migrations"
)

// Service represents the database service
type Service struct {
	db               *Database
	migrationService *migrations.MigrationService
	logger           *zap.Logger
}

// NewService creates a new database service
func NewService(config *DatabaseConfig, logger *zap.Logger) (*Service, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Create database connection
	db, err := NewDatabase(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Create migration service
	migrationService := migrations.NewMigrationService(db, logger)

	service := &Service{
		db:               db,
		migrationService: migrationService,
		logger:           logger,
	}

	return service, nil
}

// Initialize initializes the database service
func (s *Service) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing database service")

	// Run migrations
	if err := s.migrationService.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	s.logger.Info("Database service initialized successfully")
	return nil
}

// Close closes the database service
func (s *Service) Close() error {
	s.logger.Info("Closing database service")
	return s.db.Close()
}

// Health checks the health of the database service
func (s *Service) Health(ctx context.Context) error {
	return s.db.Health(ctx)
}

// GetDatabase returns the underlying database instance
func (s *Service) GetDatabase() *Database {
	return s.db
}

// GetMigrationService returns the migration service
func (s *Service) GetMigrationService() *migrations.MigrationService {
	return s.migrationService
}

// GetCurrentMigrationVersion returns the current migration version
func (s *Service) GetCurrentMigrationVersion(ctx context.Context) (int, error) {
	return s.migrationService.GetCurrentVersion(ctx)
}

// RollbackToVersion rolls back the database to a specific version
func (s *Service) RollbackToVersion(ctx context.Context, version int) error {
	return s.migrationService.Rollback(ctx, version)
}
