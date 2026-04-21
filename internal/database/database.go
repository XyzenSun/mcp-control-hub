package database

import (
	"fmt"
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
