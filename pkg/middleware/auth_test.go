package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")

	// Test regular request
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.JSON(200, gin.H{"request_id": requestID})
	})

	// Test without existing request ID
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

	// Test with existing request ID
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", "test-id-123")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test-id-123", w.Header().Get("X-Request-ID"))
}

func TestTenantMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TenantMiddleware())
	router.GET("/test", func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		c.JSON(200, gin.H{"tenant_id": tenantID})
	})

	// Test with tenant ID
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Tenant-ID", "tenant-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// Response would contain the tenant_id from context
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := config.JWTConfig{Secret: "test-secret"}

	router := gin.New()
	router.Use(RequireAuth(jwtConfig))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "authenticated"})
	})

	// Test without authorization header
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)

	// Test with invalid token
	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)

	// Test with valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   "123",
		"tenant_id": "tenant-123",
	})
	tokenString, _ := token.SignedString([]byte(jwtConfig.Secret))

	req = httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
