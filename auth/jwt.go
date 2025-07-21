// File: common/prism-common-libs/auth/jwt.go
package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	UserIDKey   = "user_id"
	ClaimsKey   = "claims"
	TenantIDKey = "tenant_id"
)

// GetTenantIDFromContext adalah helper yang digunakan oleh abstraksi DB untuk mendapatkan tenantID.
func GetTenantIDFromContext(ctx context.Context) (string, error) {
	// Gin context menggunakan Value, jadi kita ambil dari sana.
	ginCtx, ok := ctx.(*gin.Context)
	if ok {
		val, exists := ginCtx.Get(TenantIDKey)
		if !exists {
			return "", fmt.Errorf("tenant ID tidak ditemukan di dalam gin context")
		}
		tenantID, ok := val.(string)
		if !ok || tenantID == "" {
			return "", fmt.Errorf("tenant ID di context bukan string atau kosong")
		}
		return tenantID, nil
	}

	// Fallback untuk non-gin context (misalnya, gRPC)
	val := ctx.Value(TenantIDKey)
	if val == nil {
		return "", fmt.Errorf("tenant ID tidak ditemukan di dalam context")
	}
	tenantID, ok := val.(string)
	if !ok || tenantID == "" {
		return "", fmt.Errorf("tenant ID di context bukan string atau kosong")
	}
	return tenantID, nil
}

// JWTMiddleware memvalidasi token JWT dan mengekstrak informasi pengguna.
// FIX: Menerima redis.Client sebagai parameter, menghilangkan kebutuhan akan init() global.
func JWTMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format, must be Bearer token"})
			c.Abort()
			return
		}

		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret key not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			jti, exists := claims["jti"].(string)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing JTI claim"})
				c.Abort()
				return
			}

			// Gunakan klien Redis yang disuntikkan
			_, err := redisClient.Get(context.Background(), jti).Result()
			if err == nil { // Jika tidak ada error, berarti key ditemukan (token dicabut)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
			if err != redis.Nil { // Jika errornya BUKAN "key not found", berarti ada masalah server
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify token with store"})
				c.Abort()
				return
			}
			// Jika err == redis.Nil, lanjutkan (token valid)

			if userID, exists := claims["sub"]; exists {
				c.Set(UserIDKey, userID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID (sub) not found in token claims"})
				c.Abort()
				return
			}
			if tenantID, exists := claims["tid"]; exists {
				c.Set(TenantIDKey, tenantID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID (tid) not found in token claims"})
				c.Abort()
				return
			}

			c.Set(ClaimsKey, claims)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token or claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID mengekstrak ID pengguna dari konteks Gin setelah validasi middleware.
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", fmt.Errorf("user ID not found in context, middleware might be missing")
	}
	idStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user ID in context is not a string")
	}
	return idStr, nil
}

func GetTenantID(c *gin.Context) (string, error) {
	tenantID, exists := c.Get(TenantIDKey)
	if !exists {
		return "", fmt.Errorf("tenant ID not found in context, middleware might be missing or misconfigured")
	}
	idStr, ok := tenantID.(string)
	if !ok {
		return "", fmt.Errorf("tenant ID in context is not a string")
	}
	return idStr, nil
}
