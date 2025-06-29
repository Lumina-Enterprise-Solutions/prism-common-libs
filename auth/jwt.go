// common/prism-common-libs/jwt/jwt.go
package auth // <-- CHANGED from 'middleware'

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// var (
// 	redisClient *redis.Client
// )

// func init() {
// 	// Middleware ini juga butuh koneksi ke Redis.
// 	// Kita bisa inisialisasi di sini atau menerimanya sebagai parameter.
// 	// Untuk kesederhanaan, kita inisialisasi di sini.
// 	redisAddr := os.Getenv("REDIS_ADDR")
// 	if redisAddr == "" {
// 		redisAddr = "cache-redis:6379"
// 	}
// 	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
// }

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

		// JWT_SECRET_KEY akan tetap dibaca dari env karena ini adalah rahasia runtime.
		// Ini adalah salah satu pengecualian yang bisa diterima.
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

			// PERUBAHAN: Gunakan klien Redis yang di-pass sebagai parameter.
			_, err := redisClient.Get(c.Request.Context(), jti).Result() // Gunakan context dari request
			if err != redis.Nil {
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify token with store"})
					c.Abort()
					return
				}
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}

			if userID, exists := claims["sub"]; exists {
				c.Set("user_id", userID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID (sub) not found in token claims"})
				c.Abort()
				return
			}
			c.Set("claims", claims)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token or claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID tidak berubah.
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
