package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"

	"github.com/project-kgo/kim/internal/discovery"
)

// Registry implements discovery.ServiceRegistry using etcd naming/endpoints.
type Registry struct {
	client  *clientv3.Client
	manager endpoints.Manager
	ttl     time.Duration

	mu         sync.Mutex
	registered bool
	instance   discovery.ServiceInstance
	leaseID    clientv3.LeaseID
	cancel     context.CancelFunc
}

// New creates an etcd-backed Registry using an existing etcd client.
func New(cli *clientv3.Client, ttl time.Duration) *Registry {
	return &Registry{
		client: cli,
		ttl:    ttl,
	}
}

func (r *Registry) Register(ctx context.Context, instance discovery.ServiceInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.registered {
		return nil
	}

	manager, err := endpoints.NewManager(r.client, instance.Name)
	if err != nil {
		return fmt.Errorf("etcd new manager: %w", err)
	}
	r.manager = manager

	ttl := r.ttl
	if ttl <= 0 {
		ttl = 15 * time.Second
	}
	lresp, err := r.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("etcd grant lease: %w", err)
	}
	r.leaseID = lresp.ID

	keepaliveCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	kaCh, err := r.client.KeepAlive(keepaliveCtx, lresp.ID)
	if err != nil {
		cancel()
		return fmt.Errorf("etcd keepalive: %w", err)
	}
	go func() {
		for range kaCh {
		}
	}()

	key := fmt.Sprintf("%s/%s", instance.Name, instance.ID)
	ep := endpoints.Endpoint{Addr: instance.Address}
	if err := r.manager.AddEndpoint(ctx, key, ep, clientv3.WithLease(lresp.ID)); err != nil {
		cancel()
		_, _ = r.client.Revoke(ctx, lresp.ID)
		return fmt.Errorf("etcd add endpoint: %w", err)
	}

	r.instance = instance
	r.registered = true
	return nil
}

func (r *Registry) Deregister(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.registered || r.manager == nil {
		return nil
	}

	key := fmt.Sprintf("%s/%s", r.instance.Name, r.instance.ID)
	_ = r.manager.DeleteEndpoint(ctx, key)
	r.registered = false
	return nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	return nil
}
