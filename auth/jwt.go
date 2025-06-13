// common/prism-common-libs/jwt/jwt.go
package auth // <-- CHANGED from 'middleware'

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

var (
	redisClient *redis.Client
)

func init() {
	// Middleware ini juga butuh koneksi ke Redis.
	// Kita bisa inisialisasi di sini atau menerimanya sebagai parameter.
	// Untuk kesederhanaan, kita inisialisasi di sini.
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "cache-redis:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
}

// JWTMiddleware validates JWT tokens and extracts user information.
func JWTMiddleware() gin.HandlerFunc {
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
			// This is a critical server-side error
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
			// Cek ke Redis
			_, err := redisClient.Get(context.Background(), jti).Result()
			if err != redis.Nil { // Jika err BUKAN karena key tidak ada
				if err != nil {
					// Error Redis yang sebenarnya
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify token with store"})
					c.Abort()
					return
				}
				// Jika tidak ada error (isRevoked punya nilai), berarti token sudah dicabut.
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
			// Simpan user ID untuk kompatibilitas ke belakang
			if userID, exists := claims["sub"]; exists {
				c.Set("user_id", userID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID (sub) not found in token claims"})
				c.Abort()
				return
			}
			// SIMPAN SELURUH KLAIM untuk digunakan oleh middleware lain seperti RBAC
			c.Set("claims", claims)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token or claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts user ID from the Gin context after middleware validation.
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", fmt.Errorf("user ID not found in context, middleware might be missing")
	}
	idStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user ID in context is not a string")
	}
	return idStr, nil
}
