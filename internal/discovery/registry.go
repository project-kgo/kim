package discovery

import (
	"context"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// ServiceInstance identifies a server instance for registration.
type ServiceInstance struct {
	ID      string
	Name    string
	Address string
}

// NewServiceInstance creates a ServiceInstance with a randomly generated ID.
func NewServiceInstance(name, address string) (ServiceInstance, error) {
	id, err := gonanoid.New()
	if err != nil {
		return ServiceInstance{}, err
	}
	return ServiceInstance{
		ID:      id,
		Name:    name,
		Address: address,
	}, nil
}

// ServiceRegistry abstracts service registration for gRPC.
// Implementations must be safe for concurrent use.
type ServiceRegistry interface {
	Register(ctx context.Context, instance ServiceInstance) error
	Deregister(ctx context.Context) error
	Close() error
}
