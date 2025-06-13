package ginutil

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the structure of health check response
type HealthResponse struct {
	Status    string    `json:"status"`    // "healthy" or "unhealthy"
	Timestamp time.Time `json:"timestamp"` // Current timestamp
	Service   string    `json:"service"`   // Service name
	Version   string    `json:"version"`   // Service version
	Uptime    string    `json:"uptime"`    // How long the service has been running
}

// startTime holds the service start time for uptime calculation
var startTime = time.Now()

// HealthCheckHandler returns a Gin handler for health checks
// This endpoint is used by Consul to determine if the service is healthy
func HealthCheckHandler(serviceName, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Calculate uptime since service started
		uptime := time.Since(startTime)

		// Create health response
		response := HealthResponse{
			Status:    "healthy",                             // Always return healthy for basic check
			Timestamp: time.Now(),                            // Current time
			Service:   serviceName,                           // Name of the service
			Version:   version,                               // Version of the service
			Uptime:    uptime.Truncate(time.Second).String(), // Uptime rounded to seconds
		}

		// Return 200 OK with health information
		// Consul considers 2xx status codes as healthy
		c.JSON(http.StatusOK, response)
	}
}

// DetailedHealthCheckHandler returns a more comprehensive health check
// This can include database connectivity, external service checks, etc.
func DetailedHealthCheckHandler(serviceName, version string, checks ...HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Calculate uptime
		uptime := time.Since(startTime)

		// Assume service is healthy initially
		overallStatus := "healthy"
		statusCode := http.StatusOK

		// Run all health checks
		checkResults := make(map[string]interface{})
		for _, check := range checks {
			checkName, checkResult, err := check.Check()
			if err != nil {
				// If any check fails, mark overall status as unhealthy
				overallStatus = "unhealthy"
				statusCode = http.StatusServiceUnavailable
				checkResults[checkName] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
			} else {
				checkResults[checkName] = map[string]interface{}{
					"status": "healthy",
					"result": checkResult,
				}
			}
		}

		// Create comprehensive health response
		response := map[string]interface{}{
			"status":    overallStatus,
			"timestamp": time.Now(),
			"service":   serviceName,
			"version":   version,
			"uptime":    uptime.Truncate(time.Second).String(),
			"checks":    checkResults,
		}

		// Return appropriate status code
		c.JSON(statusCode, response)
	}
}

// HealthChecker interface for implementing custom health checks
type HealthChecker interface {
	Check() (checkName string, result interface{}, err error)
}

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	DB interface { // This should be your database interface (e.g., *sql.DB, *gorm.DB)
		Ping() error
	}
}

// Check implements HealthChecker interface for database
func (d *DatabaseHealthChecker) Check() (string, interface{}, error) {
	err := d.DB.Ping() // Ping the database to check connectivity
	if err != nil {
		return "database", nil, err // Return error if database is unreachable
	}
	return "database", "connected", nil // Return success if database is reachable
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	Client interface { // This should be your Redis client interface
		Ping() (string, error)
	}
}

// Check implements HealthChecker interface for Redis
func (r *RedisHealthChecker) Check() (string, interface{}, error) {
	result, err := r.Client.Ping() // Ping Redis to check connectivity
	if err != nil {
		return "redis", nil, err // Return error if Redis is unreachable
	}
	return "redis", result, nil // Return success with ping result
}

// SetupHealthRoutes sets up health check routes on the main router (root level)
// This creates /health and /health/detailed endpoints
func SetupHealthRoutes(router *gin.Engine, serviceName, version string) {
	// Basic health check endpoint - just returns 200 OK
	// This is what Consul will use by default
	router.GET("/health", HealthCheckHandler(serviceName, version))

	// Detailed health check endpoint with database and Redis checks
	// You can customize this based on your service dependencies
	// router.GET("/health/detailed", DetailedHealthCheckHandler(
	// 	serviceName,
	// 	version,
	// 	&DatabaseHealthChecker{DB: yourDatabase},
	// 	&RedisHealthChecker{Client: yourRedisClient},
	// ))
}

// SetupHealthRoutesForGroup sets up health check routes within a specific route group
// This creates /group-prefix/health and /group-prefix/health/detailed endpoints
func SetupHealthRoutesForGroup(group *gin.RouterGroup, serviceName, version string) {
	// Basic health check endpoint within the group
	group.GET("/health", HealthCheckHandler(serviceName, version))

	// Detailed health check endpoint within the group
	// Uncomment and customize based on your service dependencies
	// group.GET("/health/detailed", DetailedHealthCheckHandler(
	// 	serviceName,
	// 	version,
	// 	&DatabaseHealthChecker{DB: yourDatabase},
	// 	&RedisHealthChecker{Client: yourRedisClient},
	// ))
}

// SetupBothHealthRoutes sets up health routes at both root level and within a group
// This gives you both /health (for Consul) and /group-prefix/health (for service-specific checks)
func SetupBothHealthRoutes(router *gin.Engine, group *gin.RouterGroup, serviceName, version string) {
	// Root level health check for Consul and load balancers
	SetupHealthRoutes(router, serviceName, version)

	// Group-specific health check for service monitoring
	SetupHealthRoutesForGroup(group, serviceName, version)
}
