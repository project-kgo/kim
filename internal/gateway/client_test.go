package gateway

import (
	"testing"
	"time"

	"google.golang.org/grpc/resolver"
)

type fakeResolverBuilder struct{}

func (f *fakeResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	return nil, nil
}
func (f *fakeResolverBuilder) Scheme() string { return "fake" }

func TestNewClientRequiresGatewayService(t *testing.T) {
	if _, err := NewClient(Config{}, nil, &fakeResolverBuilder{}); err == nil {
		t.Fatal("expected error for empty gateway service")
	}
}

func TestNewClientRequiresResolverBuilder(t *testing.T) {
	if _, err := NewClient(Config{GatewayService: "kim-gate", GatewayTimeout: time.Second}, nil, nil); err == nil {
		t.Fatal("expected error for nil resolver builder")
	}
}

func TestNewClientDefaultsTimeout(t *testing.T) {
	c, err := NewClient(Config{GatewayService: "kim-gate"}, nil, &fakeResolverBuilder{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer c.Close()
	if c.Service() == nil {
		t.Fatal("service client is nil")
	}
}

func TestCloseNilSafe(t *testing.T) {
	var c *Client
	if err := c.Close(); err != nil {
		t.Fatalf("Close on nil returned error: %v", err)
	}
}
