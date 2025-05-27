package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set test environment variables - check errors
	if err := os.Setenv("ENVIRONMENT", "test"); err != nil {
		t.Fatalf("Failed to set ENVIRONMENT: %v", err)
	}
	if err := os.Setenv("DB_HOST", "test-db"); err != nil {
		t.Fatalf("Failed to set DB_HOST: %v", err)
	}
	if err := os.Setenv("DB_PORT", "5432"); err != nil {
		t.Fatalf("Failed to set DB_PORT: %v", err)
	}
	if err := os.Setenv("REDIS_HOST", "test-redis"); err != nil {
		t.Fatalf("Failed to set REDIS_HOST: %v", err)
	}
	if err := os.Setenv("JWT_SECRET", "test-secret"); err != nil {
		t.Fatalf("Failed to set JWT_SECRET: %v", err)
	}

	defer func() {
		// Clean up - check errors
		if err := os.Unsetenv("ENVIRONMENT"); err != nil {
			t.Errorf("Failed to unset ENVIRONMENT: %v", err)
		}
		if err := os.Unsetenv("DB_HOST"); err != nil {
			t.Errorf("Failed to unset DB_HOST: %v", err)
		}
		if err := os.Unsetenv("DB_PORT"); err != nil {
			t.Errorf("Failed to unset DB_PORT: %v", err)
		}
		if err := os.Unsetenv("REDIS_HOST"); err != nil {
			t.Errorf("Failed to unset REDIS_HOST: %v", err)
		}
		if err := os.Unsetenv("JWT_SECRET"); err != nil {
			t.Errorf("Failed to unset JWT_SECRET: %v", err)
		}
	}()

	config, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "test", config.Environment)
	assert.Equal(t, "test-db", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "test-redis", config.Redis.Host)
	assert.Equal(t, "test-secret", config.JWT.Secret)
}

func TestDatabaseDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "pass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestRedisAddress(t *testing.T) {
	cfg := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	address := cfg.Address()
	expected := "localhost:6379"
	assert.Equal(t, expected, address)
}

func TestGetEnvString(t *testing.T) {
	os.Setenv("TEST_STRING", "test_value")
	defer os.Unsetenv("TEST_STRING")

	value := getEnvString("TEST_STRING", "default")
	assert.Equal(t, "test_value", value)

	value = getEnvString("NON_EXISTENT", "default")
	assert.Equal(t, "default", value)
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "123")
	defer os.Unsetenv("TEST_INT")

	value := getEnvInt("TEST_INT", 456)
	assert.Equal(t, 123, value)

	value = getEnvInt("NON_EXISTENT", 456)
	assert.Equal(t, 456, value)

	// Test invalid int
	os.Setenv("INVALID_INT", "not_a_number")
	value = getEnvInt("INVALID_INT", 789)
	assert.Equal(t, 789, value)
	os.Unsetenv("INVALID_INT")
}
