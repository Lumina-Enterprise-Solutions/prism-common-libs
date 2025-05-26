package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/cache"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/database"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/models"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/utils"
	"github.com/google/uuid"
)

func BenchmarkDatabaseOperations(b *testing.B) {
	cfg, err := config.Load()
	if err != nil {
		b.Skip("Database not available for benchmarking")
	}

	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		b.Skip("Database not available for benchmarking")
	}

	// Setup
	db.DB.AutoMigrate(&models.User{})

	b.ResetTimer()

	b.Run("CreateUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := models.User{
				Email:     fmt.Sprintf("bench%d@example.com", i),
				FirstName: "Bench",
				LastName:  "User",
				Status:    "active",
			}
			db.DB.Create(&user)
		}
	})

	b.Run("FindUser", func(b *testing.B) {
		// Create test user
		user := models.User{
			Email:     "find@example.com",
			FirstName: "Find",
			LastName:  "User",
			Status:    "active",
		}
		db.DB.Create(&user)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var foundUser models.User
			db.DB.First(&foundUser, user.ID)
		}
	})

	b.Run("UpdateUser", func(b *testing.B) {
		// Create test user
		user := models.User{
			Email:     "update@example.com",
			FirstName: "Update",
			LastName:  "User",
			Status:    "active",
		}
		db.DB.Create(&user)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user.FirstName = fmt.Sprintf("Updated%d", i)
			db.DB.Save(&user)
		}
	})

	// Cleanup
	db.DB.Exec("DELETE FROM users WHERE email LIKE 'bench%@example.com' OR email IN ('find@example.com', 'update@example.com')")
}

func BenchmarkRedisOperations(b *testing.B) {
	cfg, err := config.Load()
	if err != nil {
		b.Skip("Redis not available for benchmarking")
	}

	redisClient := cache.NewRedisClient(cfg.Redis)
	ctx := context.Background()

	// Test if Redis is available
	err = redisClient.Set(ctx, "test", "test", time.Second)
	if err != nil {
		b.Skip("Redis not available for benchmarking")
	}

	testData := map[string]interface{}{
		"id":    "123",
		"name":  "Test User",
		"email": "test@example.com",
		"data":  []string{"item1", "item2", "item3"},
	}

	b.ResetTimer()

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:set:%d", i)
			redisClient.Set(ctx, key, testData, time.Hour)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Setup data
		key := "bench:get:data"
		redisClient.Set(ctx, key, testData, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result map[string]interface{}
			redisClient.Get(ctx, key, &result)
		}
	})

	b.Run("Exists", func(b *testing.B) {
		// Setup data
		key := "bench:exists:data"
		redisClient.Set(ctx, key, testData, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			redisClient.Exists(ctx, key)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:delete:%d", i)
			redisClient.Set(ctx, key, testData, time.Hour)
			b.StartTimer()
			redisClient.Delete(ctx, key)
			b.StopTimer()
		}
	})
}

func BenchmarkCryptoOperations(b *testing.B) {
	b.Run("GenerateRandomBytes16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomBytes(16)
		}
	})

	b.Run("GenerateRandomBytes32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomBytes(32)
		}
	})

	b.Run("GenerateRandomBytes64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomBytes(64)
		}
	})

	b.Run("GenerateRandomString16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomString(16)
		}
	})

	b.Run("GenerateRandomString32", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomString(32)
		}
	})

	b.Run("GenerateRandomString64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			utils.GenerateRandomString(64)
		}
	})
}

func BenchmarkUUIDGeneration(b *testing.B) {
	b.Run("UUIDv4", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuid.New()
		}
	})

	b.Run("UUIDString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuid.New().String()
		}
	})
}

func BenchmarkConfigLoad(b *testing.B) {
	b.Run("LoadConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config.Load()
		}
	})
}
