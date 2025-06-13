// file: prism-common-libs/client/consul.go
package client // Pastikan package name sudah benar

import (
	"fmt"
	"log"
	"os"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
)

// ServiceRegistrationInfo berisi semua informasi yang dibutuhkan untuk mendaftarkan service.
// Ini membuat parameter fungsi menjadi bersih dan terstruktur.
type ServiceRegistrationInfo struct {
	ServiceName    string            // Nama logis service, e.g., "prism-auth-service"
	ServiceID      string            // ID unik service, e.g., "prism-auth-service-8080"
	Port           int               // Port tempat service berjalan
	Tags           []string          // Tag untuk service discovery (termasuk tag Traefik)
	Meta           map[string]string // Metadata tambahan
	HealthCheckURL string            // URL lengkap untuk health check, e.g., "http://prism_auth_service:8080/auth/health"
}

// RegisterService mendaftarkan sebuah service ke Consul berdasarkan info yang diberikan.
func RegisterService(info ServiceRegistrationInfo) (*consulapi.Client, error) {
	// Konfigurasi koneksi ke Consul Agent
	consulConfig := consulapi.DefaultConfig()
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		consulAddr = "consul:8500" // Alamat Consul di dalam jaringan Docker
	}
	consulConfig.Address = consulAddr

	client, err := consulapi.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat consul client: %w", err)
	}

	// Buat payload registrasi
	registration := &consulapi.AgentServiceRegistration{
		ID:      info.ServiceID,
		Name:    info.ServiceName,
		Port:    info.Port,
		Tags:    info.Tags,
		Meta:    info.Meta,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           info.HealthCheckURL,
			Interval:                       "10s",
			Timeout:                        "3s",
			DeregisterCriticalServiceAfter: "30s",
			Status:                         consulapi.HealthPassing, // Mulai dengan status sehat
		},
	}

	// Daftarkan service
	if err := client.Agent().ServiceRegister(registration); err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan service '%s' ke consul: %w", info.ServiceName, err)
	}

	log.Printf("Berhasil mendaftarkan service '%s' (ID: %s) ke Consul.", info.ServiceName, info.ServiceID)
	log.Printf("Health check dikonfigurasi ke: %s", info.HealthCheckURL)

	return client, nil
}

// DeregisterService menghapus registrasi service dari Consul.
// Menerima client Consul yang sama yang dikembalikan oleh RegisterService.
func DeregisterService(client *consulapi.Client, serviceID string) {
	if client == nil {
		log.Printf("Peringatan: Mencoba menghapus registrasi dengan Consul client yang nil untuk service ID %s.", serviceID)
		return
	}
	if err := client.Agent().ServiceDeregister(serviceID); err != nil {
		log.Printf("Gagal menghapus registrasi service '%s': %v", serviceID, err)
	} else {
		log.Printf("Berhasil menghapus registrasi service '%s'", serviceID)
	}
}

// TraefikTagBuilder adalah helper untuk membangun tag Traefik secara terstruktur.
type TraefikTagBuilder struct {
	ServiceName string
	PathPrefix  string
	Middlewares []string
	Port        int
	Priority    int
}

// Build menghasilkan slice of strings yang berisi tag Traefik.
func (b *TraefikTagBuilder) Build() []string {
	tags := []string{
		"traefik.enable=true",
		fmt.Sprintf("traefik.http.routers.%s.rule=PathPrefix(`%s`)", b.ServiceName, b.PathPrefix),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%d", b.ServiceName, b.Port),
	}

	if b.Priority > 0 {
		// PERBAIKAN DI SINI: Tukar urutan argumen
		tags = append(tags, fmt.Sprintf("traefik.http.routers.%s.priority=%d", b.ServiceName, b.Priority))
	}

	// Gabungkan middleware dengan aman
	var allMiddlewares []string
	middlewareName := fmt.Sprintf("%s-stripprefix", b.ServiceName)
	tags = append(tags, fmt.Sprintf("traefik.http.middlewares.%s.stripprefix.prefixes=%s", middlewareName, b.PathPrefix))
	allMiddlewares = append(allMiddlewares, middlewareName)
	allMiddlewares = append(allMiddlewares, b.Middlewares...)

	tags = append(tags, fmt.Sprintf("traefik.http.routers.%s.middlewares=%s", b.ServiceName, strings.Join(allMiddlewares, ",")))

	return tags
}
