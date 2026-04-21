package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mcp-control-hub/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(driver, dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// For SQLite, ensure the database directory exists
	if driver == "sqlite" {
		if err := ensureSQLiteDirectory(dsn); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	switch driver {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "mysql":
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// Auto migrate models
	if err := db.AutoMigrate(
		&models.UpstreamServer{},
		&models.Tool{},
		&models.Namespace{},
		&models.APIKey{},
		&models.SystemConfig{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Add unique index for (ServerID, Name) on tools
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tools_server_name ON tools(server_id, name)").Error; err != nil {
		return nil, fmt.Errorf("failed to create unique index: %w", err)
	}

	return db, nil
}

// ensureSQLiteDirectory creates the parent directory for SQLite database file if it doesn't exist
func ensureSQLiteDirectory(dsn string) error {
	// Get the directory path from the DSN
	dir := filepath.Dir(dsn)
	// If it's just a filename (no directory), use current directory
	if dir == "." || dir == "" {
		return nil
	}
	// Create directory with 0755 permissions (rwxr-xr-x)
	return os.MkdirAll(dir, 0755)
}
