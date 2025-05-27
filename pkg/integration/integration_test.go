package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/cache"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/database"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/middleware"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	db          *database.PostgresDB
	redisClient *cache.RedisClient
	config      *config.Config
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Define test configuration to match docker-compose.test.yml
	cfg := &config.Config{
		Environment: "test",
		TenantID:    "default",
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		JWT: config.JWTConfig{
			Secret:         "test-secret-key",
			ExpirationTime: 3600,
		},
		Server: config.ServerConfig{
			Port:         8080,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
	}
	suite.config = cfg

	// Setup database
	db, err := database.NewPostgresConnection(&cfg.Database)
	require.NoError(suite.T(), err)
	suite.db = db

	// Setup Redis
	redisClient := cache.NewRedisClient(cfg.Redis)
	suite.redisClient = redisClient

	// Run migrations
	err = suite.db.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Tenant{})
	require.NoError(suite.T(), err)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	// Clean up test data
	if suite.db != nil {
		suite.db.DB.Exec("DROP TABLE IF EXISTS user_roles, users, roles, tenants CASCADE")
	}
}

func (suite *IntegrationTestSuite) SetupTest() {
	// Clean up data before each test
	suite.db.DB.Exec("DELETE FROM user_roles")
	suite.db.DB.Exec("DELETE FROM users")
	suite.db.DB.Exec("DELETE FROM roles")
	suite.db.DB.Exec("DELETE FROM tenants")
}

func (suite *IntegrationTestSuite) TestDatabaseOperations() {
	// Create a tenant
	tenant := models.Tenant{
		Name:   "Test Tenant",
		Slug:   "test_tenant",
		Status: "active",
	}

	result := suite.db.DB.Create(&tenant)
	require.NoError(suite.T(), result.Error)
	assert.NotEqual(suite.T(), uuid.Nil, tenant.ID)

	// Create a role
	role := models.Role{
		Name: "admin",
		Permissions: map[string]interface{}{
			"users": []string{"create", "read", "update", "delete"},
		},
	}

	result = suite.db.DB.Create(&role)
	require.NoError(suite.T(), result.Error)

	// Create a user
	user := models.User{
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		FirstName:    "Test",
		LastName:     "User",
		Status:       "active",
		Roles:        []models.Role{role},
	}

	result = suite.db.DB.Create(&user)
	require.NoError(suite.T(), result.Error)

	// Test retrieval with associations
	var retrievedUser models.User
	result = suite.db.DB.Preload("Roles").First(&retrievedUser, user.ID)
	require.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
	assert.Len(suite.T(), retrievedUser.Roles, 1)
	assert.Equal(suite.T(), role.Name, retrievedUser.Roles[0].Name)
}

func (suite *IntegrationTestSuite) TestRedisOperations() {
	ctx := context.Background()

	// Test basic operations
	testData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// Set data
	err := suite.redisClient.Set(ctx, "user:123", testData, time.Minute)
	require.NoError(suite.T(), err)

	// Check existence
	exists, err := suite.redisClient.Exists(ctx, "user:123")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Get data
	var retrieved map[string]interface{}
	err = suite.redisClient.Get(ctx, "user:123", &retrieved)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testData["name"], retrieved["name"])
	assert.Equal(suite.T(), testData["email"], retrieved["email"])

	// Delete data
	err = suite.redisClient.Delete(ctx, "user:123")
	require.NoError(suite.T(), err)

	// Verify deletion
	exists, err = suite.redisClient.Exists(ctx, "user:123")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *IntegrationTestSuite) TestFullAuthenticationFlow() {
	gin.SetMode(gin.TestMode)

	// Setup router with middleware
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.TenantMiddleware())

	// Public endpoint
	router.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "public"})
	})

	// Protected endpoint
	protected := router.Group("/protected")
	protected.Use(middleware.RequireAuth(suite.config.JWT))
	protected.GET("/user", func(c *gin.Context) {
		userID := c.GetString("user_id")
		tenantID := c.GetString("tenant_id")
		c.JSON(200, gin.H{
			"user_id":   userID,
			"tenant_id": tenantID,
			"message":   "authenticated",
		})
	})

	// Test public endpoint
	req := httptest.NewRequest("GET", "/public", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)

	// Test protected endpoint without auth
	req = httptest.NewRequest("GET", "/protected/user", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 401, w.Code)

	// Test protected endpoint with invalid token
	req = httptest.NewRequest("GET", "/protected/user", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 401, w.Code)

	// Test protected endpoint with valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   "user-123",
		"tenant_id": "tenant-456",
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(suite.config.JWT.Secret))
	require.NoError(suite.T(), err)

	req = httptest.NewRequest("GET", "/protected/user", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("X-Tenant-ID", "tenant-456")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(suite.T(), 200, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "user-123")
	assert.Contains(suite.T(), w.Body.String(), "tenant-456")
}

func (suite *IntegrationTestSuite) TestTenantIsolation() {
	// Create test tenants
	tenant1 := models.Tenant{Name: "Tenant 1", Slug: "tenant-1", Status: "active"}
	tenant2 := models.Tenant{Name: "Tenant 2", Slug: "tenant-2", Status: "active"}

	suite.db.DB.Create(&tenant1)
	suite.db.DB.Create(&tenant2)

	// Test tenant context switching
	tenantDB1 := suite.db.WithTenant(tenant1.ID.String())
	tenantDB2 := suite.db.WithTenant(tenant2.ID.String())

	// Both should be valid database connections
	assert.NotNil(suite.T(), tenantDB1)
	assert.NotNil(suite.T(), tenantDB2)
}

func (suite *IntegrationTestSuite) TestCacheWithDatabase() {
	ctx := context.Background()

	// Create a user in database
	user := models.User{
		Email:     "cache@example.com",
		FirstName: "Cache",
		LastName:  "User",
		Status:    "active",
	}

	result := suite.db.DB.Create(&user)
	require.NoError(suite.T(), result.Error)

	// Cache the user
	cacheKey := fmt.Sprintf("user:%s", user.ID.String())
	err := suite.redisClient.Set(ctx, cacheKey, user, time.Hour)
	require.NoError(suite.T(), err)

	// Retrieve from cache
	var cachedUser models.User
	err = suite.redisClient.Get(ctx, cacheKey, &cachedUser)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Email, cachedUser.Email)
	assert.Equal(suite.T(), user.FirstName, cachedUser.FirstName)

	// Update user in database
	user.FirstName = "Updated"
	suite.db.DB.Save(&user)

	// Cache should still have old data
	err = suite.redisClient.Get(ctx, cacheKey, &cachedUser)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cache", cachedUser.FirstName) // Old value

	// Invalidate cache
	err = suite.redisClient.Delete(ctx, cacheKey)
	require.NoError(suite.T(), err)

	// Re-cache with updated data
	err = suite.redisClient.Set(ctx, cacheKey, user, time.Hour)
	require.NoError(suite.T(), err)

	err = suite.redisClient.Get(ctx, cacheKey, &cachedUser)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", cachedUser.FirstName) // New value
}

func (suite *IntegrationTestSuite) TestComplexUserRolePermissions() {
	// Create multiple roles with different permissions
	adminRole := models.Role{
		Name: "admin",
		Permissions: map[string]interface{}{
			"users":   []string{"create", "read", "update", "delete"},
			"reports": []string{"create", "read", "update", "delete"},
			"system":  []string{"configure", "maintain"},
		},
	}

	userRole := models.Role{
		Name: "user",
		Permissions: map[string]interface{}{
			"users":   []string{"read", "update"},
			"reports": []string{"read"},
		},
	}

	suite.db.DB.Create(&adminRole)
	suite.db.DB.Create(&userRole)

	// Create user with multiple roles
	user := models.User{
		Email:     "multi-role@example.com",
		FirstName: "Multi",
		LastName:  "Role",
		Status:    "active",
		Roles:     []models.Role{adminRole, userRole},
	}

	result := suite.db.DB.Create(&user)
	require.NoError(suite.T(), result.Error)

	// Verify user has both roles
	var retrievedUser models.User
	result = suite.db.DB.Preload("Roles").First(&retrievedUser, user.ID)
	require.NoError(suite.T(), result.Error)
	assert.Len(suite.T(), retrievedUser.Roles, 2)
}

func (suite *IntegrationTestSuite) TestConcurrentCacheOperations() {
	ctx := context.Background()

	// Test concurrent writes to different keys
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			key := fmt.Sprintf("concurrent:test:%d", index)
			data := map[string]interface{}{
				"index": index,
				"data":  fmt.Sprintf("test-data-%d", index),
			}

			err := suite.redisClient.Set(ctx, key, data, time.Minute)
			assert.NoError(suite.T(), err)

			var retrieved map[string]interface{}
			err = suite.redisClient.Get(ctx, key, &retrieved)
			assert.NoError(suite.T(), err)
			assert.Equal(suite.T(), float64(index), retrieved["index"]) // JSON unmarshals numbers as float64

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func (suite *IntegrationTestSuite) TestDatabaseTransactions() {
	// Begin transaction
	tx := suite.db.DB.Begin()

	// Create tenant in transaction
	tenant := models.Tenant{
		Name:   "Transaction Test Tenant",
		Slug:   "tx-test_tenant",
		Status: "active",
	}

	result := tx.Create(&tenant)
	require.NoError(suite.T(), result.Error)

	// Verify tenant exists within transaction
	var foundTenant models.Tenant
	result = tx.First(&foundTenant, tenant.ID)
	require.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), tenant.Name, foundTenant.Name)

	// Rollback transaction
	tx.Rollback()

	// Verify tenant doesn't exist after rollback
	result = suite.db.DB.First(&foundTenant, tenant.ID)
	assert.Error(suite.T(), result.Error) // Should not find the tenant
}

func (suite *IntegrationTestSuite) TestMiddlewareChaining() {
	gin.SetMode(gin.TestMode)

	// Setup router with full middleware chain
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.TenantMiddleware())

	// Add test endpoint to verify all middleware is working
	router.GET("/test-middleware", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		tenantID := c.GetString("tenant_id")

		c.JSON(200, gin.H{
			"request_id": requestID,
			"tenant_id":  tenantID,
			"cors":       c.GetHeader("Access-Control-Allow-Origin"),
		})
	})

	req := httptest.NewRequest("GET", "/test-middleware", http.NoBody)
	req.Header.Set("X-Tenant-ID", "test_tenant-123")
	req.Header.Set("X-Request-ID", "custom-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), 200, w.Code)
	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(suite.T(), "custom-request-id", w.Header().Get("X-Request-ID"))
	assert.Contains(suite.T(), w.Body.String(), "test_tenant-123")
}

func (suite *IntegrationTestSuite) TestCacheEvictionAndExpiry() {
	ctx := context.Background()

	// Test with short expiry
	testData := map[string]string{"key": "value"}
	err := suite.redisClient.Set(ctx, "expiry:test", testData, time.Millisecond*100)
	require.NoError(suite.T(), err)

	// Verify data exists immediately
	exists, err := suite.redisClient.Exists(ctx, "expiry:test")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Wait for expiry
	time.Sleep(time.Millisecond * 150)

	// Verify data has expired
	exists, err = suite.redisClient.Exists(ctx, "expiry:test")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
