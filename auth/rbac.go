package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	userv1 "github.com/Lumina-Enterprise-Solutions/prism-protobufs/gen/go/prism/user/v1"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// permissionCache menyimpan izin untuk setiap peran dengan TTL.
type permissionCache struct {
	mu          sync.RWMutex
	permissions map[string]cachedPermissions
}

type cachedPermissions struct {
	permissions map[string]struct{} // Menggunakan map kosong untuk lookup O(1)
	expiresAt   time.Time
}

// RBACMiddleware adalah middleware yang stateful dengan koneksi gRPC dan cache.
type RBACMiddleware struct {
	userSvcClient userv1.UserServiceClient
	cache         *permissionCache
	cacheTTL      time.Duration
}

// NewRBACMiddleware membuat instance RBAC middleware baru.
func NewRBACMiddleware(userServiceAddress string, cacheTTL time.Duration) (*RBACMiddleware, error) {
	conn, err := grpc.NewClient(userServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to user service for RBAC: %w", err)
	}

	return &RBACMiddleware{
		userSvcClient: userv1.NewUserServiceClient(conn),
		cache: &permissionCache{
			permissions: make(map[string]cachedPermissions),
		},
		cacheTTL: cacheTTL,
	}, nil
}

// getPermissionsForRole mengambil izin untuk sebuah peran, menggunakan cache jika memungkinkan.
func (m *RBACMiddleware) getPermissionsForRole(roleName string) (map[string]struct{}, error) {
	m.cache.mu.RLock()
	cached, found := m.cache.permissions[roleName]
	m.cache.mu.RUnlock()

	if found && time.Now().Before(cached.expiresAt) {
		return cached.permissions, nil
	}

	// Jika tidak ada di cache atau sudah kedaluwarsa, ambil dari user-service
	req := &userv1.GetPermissionsForRoleRequest{RoleName: roleName}
	resp, err := m.userSvcClient.GetPermissionsForRole(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch permissions for role '%s': %w", roleName, err)
	}

	// Bangun set izin untuk lookup yang cepat
	permsSet := make(map[string]struct{})
	for _, p := range resp.GetPermissions() {
		permsSet[p] = struct{}{}
	}

	// Simpan ke cache
	m.cache.mu.Lock()
	m.cache.permissions[roleName] = cachedPermissions{
		permissions: permsSet,
		expiresAt:   time.Now().Add(m.cacheTTL),
	}
	m.cache.mu.Unlock()

	return permsSet, nil
}

// RequirePermission membuat middleware yang memeriksa apakah pengguna memiliki izin yang diperlukan.
func (m *RBACMiddleware) RequirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Claims not found in context"})
			c.Abort()
			return
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}

		role, ok := mapClaims["role"].(string)
		if !ok || role == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		// Ambil izin untuk peran ini (dari cache atau gRPC)
		userPermissions, err := m.getPermissionsForRole(role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify permissions", "details": err.Error()})
			c.Abort()
			return
		}

		// Periksa apakah izin yang dibutuhkan ada
		if _, hasPermission := userPermissions[requiredPermission]; !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Access denied. Required permission: '%s'", requiredPermission)})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly adalah shortcut untuk memeriksa peran 'admin'. Ini berguna untuk akses super.
// Untuk kontrol yang lebih halus, gunakan RequirePermission.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Claims not found in context"})
			c.Abort()
			return
		}
		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}
		role, ok := mapClaims["role"].(string)
		if !ok || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Administrator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
