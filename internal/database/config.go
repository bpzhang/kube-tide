package database

import (
	"fmt"
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            DatabaseType  `mapstructure:"type" validate:"required,oneof=postgres sqlite"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database" validate:"required"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`

	// SQLite specific settings
	SQLiteFilePath string `mapstructure:"sqlite_file_path"`
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            PostgreSQL,
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Database:        "kube_tide",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		SQLiteFilePath:  "./data/kube_tide.db",
	}
}

// DSN returns the data source name for the database
func (c *DatabaseConfig) DSN() string {
	switch c.Type {
	case PostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
	case SQLite:
		return c.SQLiteFilePath
	default:
		return ""
	}
}

// DriverName returns the driver name for the database type
func (c *DatabaseConfig) DriverName() string {
	switch c.Type {
	case PostgreSQL:
		return "postgres"
	case SQLite:
		return "sqlite"
	default:
		return ""
	}
}
