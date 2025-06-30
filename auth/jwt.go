// File: prism-common-libs/auth/jwt.go
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

// UserIDKey is the key used in the Gin context to store the user's ID.
const UserIDKey = "user_id"

// ClaimsKey is the key used in the Gin context to store the full JWT claims map.
const ClaimsKey = "claims"

var (
	redisClient *redis.Client
)

func init() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "cache-redis:6379"
	}
	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
}

// JWTMiddleware validates JWT tokens and extracts user information.
// FIX: The function signature is now parameter-less as it initializes its own Redis client.
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
			_, err := redisClient.Get(context.Background(), jti).Result()
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
				// FIX: Use the exported constant for the context key.
				c.Set(UserIDKey, userID)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID (sub) not found in token claims"})
				c.Abort()
				return
			}
			// FIX: Use the exported constant for the context key.
			c.Set(ClaimsKey, claims)
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
	// FIX: Use the exported constant for the context key.
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
