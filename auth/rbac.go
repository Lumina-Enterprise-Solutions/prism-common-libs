// file: common/prism-common-libs/rbac/rbac.go
package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// Definisikan peran sebagai konstanta untuk menghindari typo
	AdminRole = "admin"
	UserRole  = "user"
)

// Helper untuk mengambil klaim dari context Gin
func getClaims(c *gin.Context) (jwt.MapClaims, error) {
	// Asumsi: commonjwt.JWTMiddleware() sudah menaruh token yang ter-parse
	// di context, tapi mari kita buat ini lebih robust dengan mengambilnya langsung.
	// Untuk sekarang, kita akan tambahkan claims ke context di middleware JWT.

	claims, exists := c.Get("claims")
	if !exists {
		return nil, fmt.Errorf("claims tidak ditemukan di context, pastikan JWTMiddleware berjalan sebelumnya")
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims di context bukan tipe jwt.MapClaims")
	}
	return mapClaims, nil
}

// RequireRole adalah middleware factory yang memeriksa peran tertentu.
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getClaims(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal memproses klaim token", "details": err.Error()})
			c.Abort()
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Klaim 'role' tidak ditemukan atau bukan string di dalam token"})
			c.Abort()
			return
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Akses ditolak. Membutuhkan peran '%s'", requiredRole)})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly adalah shortcut untuk RequireRole("admin")
func AdminOnly() gin.HandlerFunc {
	return RequireRole(AdminRole)
}
