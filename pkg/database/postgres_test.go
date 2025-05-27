package database

import (
	"testing"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresConnection(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     543,
		Database: "test_db",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	db, err := NewPostgresConnection(&cfg)
	if err != nil {
		t.Skipf("Skipping database test: %v", err)
		return
	}

	require.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, db.DB)

	// Test connection
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)
	err = sqlDB.Ping()
	assert.NoError(t, err)
}

func TestWithTenant(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "test_db",
		Username: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	db, err := NewPostgresConnection(&cfg)
	if err != nil {
		t.Skipf("Skipping database test: %v", err)
		return
	}

	require.NoError(t, err)

	tenantDB := db.WithTenant("test_tenant")
	assert.NotNil(t, tenantDB)
}
