package database

import (
	"fmt"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDB struct {
	DB *gorm.DB
}

func NewPostgresConnection(cfg config.DatabaseConfig) (*PostgresDB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) WithTenant(tenantID string) *gorm.DB {
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	return p.DB.Exec(fmt.Sprintf("SET search_path TO %s", schemaName))
}
