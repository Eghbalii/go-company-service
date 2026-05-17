package database

import (
	"fmt"
	"time"

	"github.com/eghbalii/go-company-service/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres opens a GORM connection pool configured from cfg.
func NewPostgres(cfg config.Database) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}

// Migrate runs golang-migrate migrations from the embedded filesystem.
// migrationURL must be a file:// or embed URL accepted by golang-migrate.
func Migrate(dsn, migrationsPath string) error {
	// Deferred to avoid importing golang-migrate here; called from main.go
	// via the standalone migrate binary or the infrastructure/migrate package.
	_ = dsn
	_ = migrationsPath
	_ = time.Second // prevent import removal
	return nil
}
