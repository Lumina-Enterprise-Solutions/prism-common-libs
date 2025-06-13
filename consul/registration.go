package consul

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// getLocalIP attempts to get the container's IP address in the Docker network
func getLocalIP() (string, error) {
	// Try to get IP from environment variable first (more reliable in Docker)
	if ip := os.Getenv("SERVICE_IP"); ip != "" {
		return ip, nil
	}

	// Fallback: try to detect IP by connecting to Consul
	conn, err := net.Dial("udp", "consul:8500")
	if err != nil {
		return "", fmt.Errorf("failed to connect to consul for IP detection: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// TraefikConfig holds Traefik-specific configuration for service registration
type TraefikConfig struct {
	PathPrefix   string   // e.g., "/auth", "/users"
	Middlewares  []string // e.g., ["security-headers", "cors-policy"]
	StripPrefix  bool     // Whether to strip the path prefix
	LoadBalancer string   // Load balancing method: "round_robin", "least_conn", etc.
	HealthCheck  string   // Health check path, defaults to "/health"
	Priority     int      // Router priority, higher = more priority
}

func RegisterServiceWithConsul(serviceName string, port int) {
	RegisterServiceWithTraefikConfig(serviceName, port, TraefikConfig{
		PathPrefix:   fmt.Sprintf("/%s", serviceName),
		Middlewares:  []string{"security-headers", "cors-policy", "global-rate-limit"},
		StripPrefix:  true,
		LoadBalancer: "round_robin",
		HealthCheck:  "/health",
		Priority:     100,
	})
}

func RegisterServiceWithTraefikConfig(serviceName string, port int, config TraefikConfig) {
	// Configure Consul client to connect to the Consul server
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = "consul:8500" // Explicitly set Consul address

	consul, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("Failed to create consul client: %v", err)
	}

	// Wait for Consul to be available
	log.Printf("Waiting for Consul to be available...")
	for i := 0; i < 30; i++ {
		_, err := consul.Status().Leader()
		if err == nil {
			log.Printf("Consul is available")
			break
		}
		if i == 29 {
			log.Fatalf("Consul not available after 30 attempts: %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	// Get the service IP address
	serviceIP, err := getLocalIP()
	if err != nil {
		log.Printf("Warning: Could not detect service IP, using container hostname: %v", err)
		serviceIP = "" // Let Consul agent determine the IP
	}

	serviceID := fmt.Sprintf("%s-%d", serviceName, port)

	// Build Traefik tags for service discovery
	tags := buildTraefikTags(serviceName, port, serviceIP, config)

	// FIXED: Use container name for health check instead of IP
	// This allows Consul to reach the service via Docker's internal DNS
	healthCheckTarget := getHealthCheckTarget(serviceName)

	registration := &consulapi.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Port:    port,
		Address: serviceIP, // Set the detected IP address for service discovery
		Tags:    tags,
		Meta: map[string]string{
			"version":     "1.0.0",
			"environment": getEnvironment(),
			"traefik":     "enabled",
		},
		Check: &consulapi.AgentServiceCheck{
			// FIXED: Use container name instead of IP for health check
			// Docker's internal DNS will resolve the container name to the correct IP
			HTTP:                           fmt.Sprintf("http://%s:%d%s", healthCheckTarget, port, config.HealthCheck),
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
			// ADDED: Set check status to passing initially to avoid immediate failures
			Status: "passing",
		},
	}

	// Register the service
	if err := consul.Agent().ServiceRegister(registration); err != nil {
		log.Fatalf("Failed to register service '%s' with consul: %v", serviceName, err)
	}

	log.Printf("Successfully registered service '%s' (ID: %s) with Consul at %s:%d",
		serviceName, serviceID, serviceIP, port)
	log.Printf("Health check URL: http://%s:%d%s", healthCheckTarget, port, config.HealthCheck)
	log.Printf("Traefik configuration: PathPrefix=%s, Middlewares=%v, Priority=%d",
		config.PathPrefix, config.Middlewares, config.Priority)

	// Verify registration
	services, err := consul.Agent().Services()
	if err != nil {
		log.Printf("Warning: Could not verify service registration: %v", err)
	} else {
		if service, exists := services[serviceID]; exists {
			log.Printf("✅ Service registration verified: %s at %s:%d",
				service.Service, service.Address, service.Port)
			log.Printf("✅ Traefik tags: %v", service.Tags)
		} else {
			log.Printf("⚠️ Service registration not found in agent services list")
		}
	}
}

// buildTraefikTags creates Traefik-specific tags for service discovery
func buildTraefikTags(serviceName string, port int, serviceIP string, config TraefikConfig) []string {
	tags := []string{
		// Basic service info
		"version=1.0.0",
		"environment=" + getEnvironment(),
		fmt.Sprintf("port=%d", port),

		// Enable Traefik
		"traefik.enable=true",

		// HTTP Router configuration
		fmt.Sprintf("traefik.http.routers.%s.rule=PathPrefix(`%s`)", serviceName, config.PathPrefix),
		fmt.Sprintf("traefik.http.routers.%s.priority=%d", serviceName, config.Priority),

		// Service configuration
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", serviceName, port),
	}

	// Add middlewares if specified
	if len(config.Middlewares) > 0 {
		middlewareChain := ""
		for i, middleware := range config.Middlewares {
			if i > 0 {
				middlewareChain += ","
			}
			middlewareChain += middleware
		}
		tags = append(tags, fmt.Sprintf("traefik.http.routers.%s.middlewares=%s", serviceName, middlewareChain))
	}

	// Add path prefix stripping if enabled
	if config.StripPrefix {
		tags = append(tags,
			fmt.Sprintf("traefik.http.middlewares.%s-stripprefix.stripprefix.prefixes=%s", serviceName, config.PathPrefix),
			fmt.Sprintf("traefik.http.routers.%s.middlewares=%s-stripprefix,%s", serviceName, serviceName, getMiddlewareChain(config.Middlewares)),
		)
	}

	// Add load balancing method if specified
	if config.LoadBalancer != "" {
		tags = append(tags, fmt.Sprintf("traefik.http.services.%s.loadbalancer.method=%s", serviceName, config.LoadBalancer))
	}

	// Add health check configuration
	if config.HealthCheck != "" {
		tags = append(tags,
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.path=%s", serviceName, config.HealthCheck),
			fmt.Sprintf("traefik.http.services.%s.loadbalancer.healthcheck.interval=10s", serviceName),
		)
	}

	return tags
}

// getMiddlewareChain combines custom middlewares with stripprefix
func getMiddlewareChain(middlewares []string) string {
	chain := ""
	for i, middleware := range middlewares {
		if i > 0 {
			chain += ","
		}
		chain += middleware
	}
	return chain
}

// getEnvironment returns the current environment
func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("GO_ENV"); env != "" {
		return env
	}
	return "development"
}

// FIXED: getHealthCheckTarget now returns the container name instead of IP
// This allows Consul to use Docker's internal DNS resolution
func getHealthCheckTarget(serviceName string) string {
	// Map service names to their container names as defined in docker-compose.yml
	containerNames := map[string]string{
		"prism-auth-service": "prism_auth_service",
		"prism-user-service": "prism_user_service",
	}

	// If we have a specific container name mapping, use it
	if containerName, exists := containerNames[serviceName]; exists {
		return containerName
	}

	// Otherwise, try to construct the container name from service name
	// This assumes your container naming follows the pattern: prism_<service-name>
	return fmt.Sprintf("prism_%s", serviceName)
}

// RegisterAuthService registers auth service with specific Traefik configuration
func RegisterAuthService(port int) {
	config := TraefikConfig{
		PathPrefix:   "/auth",
		Middlewares:  []string{"security-headers", "cors-policy", "login-rate-limit"},
		StripPrefix:  true,
		LoadBalancer: "round_robin",
		HealthCheck:  "/auth/health", // CHANGED: Use group-specific health check path
		Priority:     200,            // Higher priority for auth service
	}
	RegisterServiceWithTraefikConfig("prism-auth-service", port, config)
}

// RegisterUserService registers user service with specific Traefik configuration
func RegisterUserService(port int) {
	config := TraefikConfig{
		PathPrefix:   "/users",
		Middlewares:  []string{"security-headers", "cors-policy", "global-rate-limit"},
		StripPrefix:  true,
		LoadBalancer: "round_robin",
		HealthCheck:  "/users/health", // CHANGED: Use group-specific health check path
		Priority:     150,
	}
	RegisterServiceWithTraefikConfig("prism-user-service", port, config)
}

// RegisterAPIGatewayService registers a generic API service
func RegisterAPIGatewayService(serviceName string, port int, pathPrefix string, middlewares []string) {
	config := TraefikConfig{
		PathPrefix:   pathPrefix,
		Middlewares:  middlewares,
		StripPrefix:  true,
		LoadBalancer: "round_robin",
		HealthCheck:  "/health",
		Priority:     100,
	}
	RegisterServiceWithTraefikConfig(serviceName, port, config)
}

// DeregisterServiceFromConsul removes the service from Consul when shutting down
func DeregisterServiceFromConsul(serviceName string, port int) {
	config := consulapi.DefaultConfig()
	config.Address = "consul:8500"

	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Printf("Failed to create consul client for deregistration: %v", err)
		return
	}

	serviceID := fmt.Sprintf("%s-%d", serviceName, port)

	if err := consul.Agent().ServiceDeregister(serviceID); err != nil {
		log.Printf("Failed to deregister service '%s': %v", serviceID, err)
	} else {
		log.Printf("Successfully deregistered service '%s'", serviceID)
	}
}

// Helper function to register service with rate limiting for specific endpoints
func RegisterServiceWithEndpointRateLimit(serviceName string, port int, pathPrefix string, rateLimitedPaths []string) {
	middlewares := []string{"security-headers", "cors-policy", "global-rate-limit"}

	// Add specific rate limiting for certain paths
	for _, path := range rateLimitedPaths {
		if path == "/login" || path == "/register" {
			middlewares = append(middlewares, "login-rate-limit")
			break
		}
	}

	config := TraefikConfig{
		PathPrefix:   pathPrefix,
		Middlewares:  middlewares,
		StripPrefix:  true,
		LoadBalancer: "round_robin",
		HealthCheck:  "/health",
		Priority:     100,
	}
	RegisterServiceWithTraefikConfig(serviceName, port, config)
}
