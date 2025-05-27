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
	"github.com/google/uuid"
)

// Helper function to check errors in benchmarks
func checkErr(b *testing.B, err error) {
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkDatabaseOperations(b *testing.B) {
	cfg, err := config.Load()
	checkErr(b, err)
	if err != nil {
		b.Skip("Database not available for benchmarking")
	}

	db, err := database.NewPostgresConnection(&cfg.Database) // Pass pointer
	checkErr(b, err)
	if err != nil {
		b.Skip("Database not available for benchmarking")
	}

	// Setup
	err = db.DB.AutoMigrate(&models.User{})
	checkErr(b, err)

	b.ResetTimer()

	b.Run("CreateUser", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			user := models.User{
				Email:     fmt.Sprintf("bench%d@example.com", i),
				FirstName: "Bench",
				LastName:  "User",
				Status:    "active",
			}
			result := db.DB.Create(&user)
			checkErr(b, result.Error)
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
		result := db.DB.Create(&user)
		checkErr(b, result.Error)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var foundUser models.User
			result := db.DB.First(&foundUser, user.ID)
			checkErr(b, result.Error)
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
		result := db.DB.Create(&user)
		checkErr(b, result.Error)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user.FirstName = fmt.Sprintf("Updated%d", i)
			result := db.DB.Save(&user)
			checkErr(b, result.Error)
		}
	})

	// Cleanup
	result := db.DB.Exec("DELETE FROM users WHERE email LIKE 'bench%@example.com' OR email IN ('find@example.com', 'update@example.com')")
	checkErr(b, result.Error)
}

func BenchmarkRedisOperations(b *testing.B) {
	cfg, err := config.Load()
	checkErr(b, err)
	if err != nil {
		b.Skip("Redis not available for benchmarking")
	}

	redisClient := cache.NewRedisClient(cfg.Redis)
	ctx := context.Background()

	// Test if Redis is available
	err = redisClient.Set(ctx, "test", "test", time.Second)
	checkErr(b, err)
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
			err := redisClient.Set(ctx, key, testData, time.Hour)
			checkErr(b, err)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Setup data
		key := "bench:get:data"
		err := redisClient.Set(ctx, key, testData, time.Hour)
		checkErr(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result map[string]interface{}
			err := redisClient.Get(ctx, key, &result)
			checkErr(b, err)
		}
	})

	b.Run("Exists", func(b *testing.B) {
		// Setup data
		key := "bench:exists:data"
		err := redisClient.Set(ctx, key, testData, time.Hour)
		checkErr(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := redisClient.Exists(ctx, key)
			checkErr(b, err)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:delete:%d", i)
			err := redisClient.Set(ctx, key, testData, time.Hour)
			checkErr(b, err)
			b.StartTimer()
			err = redisClient.Delete(ctx, key)
			checkErr(b, err)
			b.StopTimer()
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
			_ = uuid.New().String() // Assign to _ to avoid unused result warning
		}
	})
}

func BenchmarkConfigLoad(b *testing.B) {
	b.Run("LoadConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := config.Load()
			checkErr(b, err)
		}
	})
}
