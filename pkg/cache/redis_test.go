package cache

import (
	"context"
	"testing"
	"time"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewRedisClient(t *testing.T) {
	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	client := NewRedisClient(cfg)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestRedisOperations(t *testing.T) {
	cfg := config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	client := NewRedisClient(cfg)
	ctx := context.Background()

	// Test ping first to see if Redis is available
	err := client.client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Skipping Redis test: %v", err)
		return
	}

	// Test Set and Get
	testData := TestData{Name: "test", Value: 123}
	err = client.Set(ctx, "test:key", testData, time.Minute)
	require.NoError(t, err)

	var retrieved TestData
	err = client.Get(ctx, "test:key", &retrieved)
	require.NoError(t, err)
	assert.Equal(t, testData, retrieved)

	// Test Exists
	exists, err := client.Exists(ctx, "test:key")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	err = client.Delete(ctx, "test:key")
	require.NoError(t, err)

	exists, err = client.Exists(ctx, "test:key")
	require.NoError(t, err)
	assert.False(t, exists)
}
