package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite"
)

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetDatabaseType() string
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	UpSQL       string
	DownSQL     string
}

// MigrationService handles database migrations
type MigrationService struct {
	db         DatabaseInterface
	logger     *zap.Logger
	migrations []Migration
}

// NewMigrationService creates a new migration service
func NewMigrationService(db DatabaseInterface, logger *zap.Logger) *MigrationService {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &MigrationService{
		db:         db,
		logger:     logger,
		migrations: getAllMigrations(),
	}
}

// Migrate runs all pending migrations
func (m *MigrationService) Migrate(ctx context.Context) error {
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	m.logger.Info("Starting database migration",
		zap.Int("current_version", currentVersion),
		zap.Int("target_version", m.getLatestVersion()))

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if migration.Version > currentVersion {
			if err := m.runMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
		}
	}

	m.logger.Info("Database migration completed successfully")
	return nil
}

// Rollback rolls back to a specific version
func (m *MigrationService) Rollback(ctx context.Context, targetVersion int) error {
	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if targetVersion >= currentVersion {
		return fmt.Errorf("target version %d must be less than current version %d", targetVersion, currentVersion)
	}

	m.logger.Info("Starting database rollback",
		zap.Int("current_version", currentVersion),
		zap.Int("target_version", targetVersion))

	// Sort migrations by version in descending order for rollback
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version > m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if migration.Version > targetVersion && migration.Version <= currentVersion {
			if err := m.rollbackMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
			}
		}
	}

	m.logger.Info("Database rollback completed successfully")
	return nil
}

// GetCurrentVersion returns the current migration version
func (m *MigrationService) GetCurrentVersion(ctx context.Context) (int, error) {
	return m.getCurrentVersion(ctx)
}

// createMigrationsTable creates the migrations tracking table
func (m *MigrationService) createMigrationsTable(ctx context.Context) error {
	var createTableSQL string

	switch m.db.GetDatabaseType() {
	case "postgres":
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				description TEXT NOT NULL,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			)`
	case "sqlite":
		createTableSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				description TEXT NOT NULL,
				applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`
	default:
		return fmt.Errorf("unsupported database type: %s", m.db.GetDatabaseType())
	}

	_, err := m.db.ExecContext(ctx, createTableSQL)
	return err
}

// getCurrentVersion gets the current migration version
func (m *MigrationService) getCurrentVersion(ctx context.Context) (int, error) {
	var version int
	err := m.db.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// getLatestVersion returns the latest available migration version
func (m *MigrationService) getLatestVersion() int {
	maxVersion := 0
	for _, migration := range m.migrations {
		if migration.Version > maxVersion {
			maxVersion = migration.Version
		}
	}
	return maxVersion
}

// runMigration executes a migration
func (m *MigrationService) runMigration(ctx context.Context, migration Migration) error {
	m.logger.Info("Running migration",
		zap.Int("version", migration.Version),
		zap.String("description", migration.Description))

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	_, err = tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)",
		migration.Version, migration.Description, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	m.logger.Info("Migration completed",
		zap.Int("version", migration.Version))

	return nil
}

// rollbackMigration rolls back a migration
func (m *MigrationService) rollbackMigration(ctx context.Context, migration Migration) error {
	m.logger.Info("Rolling back migration",
		zap.Int("version", migration.Version),
		zap.String("description", migration.Description))

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute rollback SQL
	if migration.DownSQL != "" {
		if _, err := tx.ExecContext(ctx, migration.DownSQL); err != nil {
			return fmt.Errorf("failed to execute rollback SQL: %w", err)
		}
	}

	// Remove migration record
	_, err = tx.ExecContext(ctx,
		"DELETE FROM schema_migrations WHERE version = $1",
		migration.Version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	m.logger.Info("Migration rollback completed",
		zap.Int("version", migration.Version))

	return nil
}
