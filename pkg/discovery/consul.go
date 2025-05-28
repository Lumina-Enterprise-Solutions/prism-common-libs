package discovery

import (
	"fmt"

	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/config"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/pkg/logger"
	"github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *api.Client
}

func NewConsulClient(cfg *config.Config) (*ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8500" // Update for production
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Consul client: %w", err)
	}
	return &ConsulClient{client: client}, nil
}

func (c *ConsulClient) RegisterService(serviceID, serviceName, address string, port int) error {
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}
	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	logger.Info("Service registered with Consul", "service_id", serviceID, "service_name", serviceName)
	return nil
}

func (c *ConsulClient) DiscoverService(serviceName string) ([]*api.ServiceEntry, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service: %w", err)
	}
	return services, nil
}
