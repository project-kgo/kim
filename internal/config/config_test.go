package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
	if cfg.RoutePrefix != DefaultRoutePrefix {
		t.Fatalf("RoutePrefix = %q, want %q", cfg.RoutePrefix, DefaultRoutePrefix)
	}
	if cfg.GRPCAddr != DefaultGRPCAddr {
		t.Fatalf("GRPCAddr = %q, want %q", cfg.GRPCAddr, DefaultGRPCAddr)
	}
	if cfg.ETCDServiceName != DefaultETCDServiceName {
		t.Fatalf("ETCDServiceName = %q, want %q", cfg.ETCDServiceName, DefaultETCDServiceName)
	}
	if cfg.GatewayService != DefaultGatewayService {
		t.Fatalf("GatewayService = %q, want %q", cfg.GatewayService, DefaultGatewayService)
	}
	if cfg.GatewayTimeout != DefaultGatewayTimeout {
		t.Fatalf("GatewayTimeout = %s, want %s", cfg.GatewayTimeout, DefaultGatewayTimeout)
	}
	if cfg.ETCDEndpointsStr != DefaultETCDEndpoints {
		t.Fatalf("ETCDEndpointsStr = %q, want %q", cfg.ETCDEndpointsStr, DefaultETCDEndpoints)
	}
	if len(cfg.ETCDEndpoints) != 1 || cfg.ETCDEndpoints[0] != "localhost:2379" {
		t.Fatalf("ETCDEndpoints = %v, want [localhost:2379]", cfg.ETCDEndpoints)
	}
	if cfg.RedisDSN != DefaultRedisDSN {
		t.Fatalf("RedisDSN = %q, want %q", cfg.RedisDSN, DefaultRedisDSN)
	}
	if cfg.DBDSN != DefaultDBDSN {
		t.Fatalf("DBDSN = %q, want %q", cfg.DBDSN, DefaultDBDSN)
	}
	if cfg.ETCDTTL != DefaultETCDTTL {
		t.Fatalf("ETCDTTL = %s, want %s", cfg.ETCDTTL, DefaultETCDTTL)
	}
}

func TestLoadYAMLConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(`
http:
  addr: ":9999"
  route_prefix: "/chat"
grpc:
  addr: ":9091"
  service: "kim-staging"
  gateway_service: "my-kim-gate"
  gateway_timeout: "10s"
etcd:
  endpoints: "host1:2379,host2:2379"
  username: "test"
  password: "secret"
  ttl: "30s"
redis:
  dsn: "redis://cache.example.com:6380/2"
db:
  dsn: "postgres://user:pass@db.example.com:5432/kim"
env: "staging"
shutdown:
  timeout: "30s"
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load([]string{"-config", path})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTPAddr != ":9999" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.RoutePrefix != "/chat" {
		t.Fatalf("RoutePrefix = %q", cfg.RoutePrefix)
	}
	if cfg.GRPCAddr != ":9091" {
		t.Fatalf("GRPCAddr = %q", cfg.GRPCAddr)
	}
	if cfg.ETCDServiceName != "kim-staging" {
		t.Fatalf("ETCDServiceName = %q", cfg.ETCDServiceName)
	}
	if cfg.GatewayService != "my-kim-gate" {
		t.Fatalf("GatewayService = %q", cfg.GatewayService)
	}
	if cfg.GatewayTimeout != 10*time.Second {
		t.Fatalf("GatewayTimeout = %s", cfg.GatewayTimeout)
	}
	if len(cfg.ETCDEndpoints) != 2 || cfg.ETCDEndpoints[0] != "host1:2379" || cfg.ETCDEndpoints[1] != "host2:2379" {
		t.Fatalf("ETCDEndpoints = %v", cfg.ETCDEndpoints)
	}
	if cfg.ETCDUsername != "test" {
		t.Fatalf("ETCDUsername = %q", cfg.ETCDUsername)
	}
	if cfg.ETCDPassword != "secret" {
		t.Fatalf("ETCDPassword = %q", cfg.ETCDPassword)
	}
	if cfg.ETCDTTL != 30*time.Second {
		t.Fatalf("ETCDTTL = %s", cfg.ETCDTTL)
	}
	if cfg.RedisDSN != "redis://cache.example.com:6380/2" {
		t.Fatalf("RedisDSN = %q", cfg.RedisDSN)
	}
	if cfg.DBDSN != "postgres://user:pass@db.example.com:5432/kim" {
		t.Fatalf("DBDSN = %q", cfg.DBDSN)
	}
	if cfg.Env != "staging" {
		t.Fatalf("Env = %q", cfg.Env)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Fatalf("ShutdownTimeout = %s", cfg.ShutdownTimeout)
	}
}

func TestLoadEnvAndFlagOverride(t *testing.T) {
	t.Setenv("KIM_HTTP_ADDR", ":9999")
	t.Setenv("KIM_ETCD_ENDPOINTS", "env1:2379,env2:2379")
	t.Setenv("KIM_GATEWAY_SERVICE", "env-gate")
	t.Setenv("KIM_GATEWAY_TIMEOUT", "15s")
	t.Setenv("KIM_REDIS_DSN", "redis://env.example.com:6379/1")
	t.Setenv("KIM_DB_DSN", "postgres://env:pass@env.example.com:5432/kim")
	t.Setenv("KIM_SHUTDOWN_TIMEOUT", "5s")

	cfg, err := Load([]string{
		"-http-addr", ":7777",
		"-etcd-endpoints", "flag1:2379",
		"-gateway-service", "flag-gate",
		"-redis-dsn", "redis://flag.example.com:6379/2",
	})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTPAddr != ":7777" {
		t.Fatalf("HTTPAddr = %q, want :7777 (flag should override env)", cfg.HTTPAddr)
	}
	if cfg.DBDSN != "postgres://env:pass@env.example.com:5432/kim" {
		t.Fatalf("DBDSN = %q, want env value", cfg.DBDSN)
	}
	if cfg.ETCDEndpointsStr != "flag1:2379" {
		t.Fatalf("ETCDEndpointsStr = %q, want flag1:2379", cfg.ETCDEndpointsStr)
	}
	if cfg.GatewayService != "flag-gate" {
		t.Fatalf("GatewayService = %q, want flag-gate", cfg.GatewayService)
	}
	if cfg.GatewayTimeout != 15*time.Second {
		t.Fatalf("GatewayTimeout = %s, want 15s (env)", cfg.GatewayTimeout)
	}
}

func TestRoutePrefixNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"kim", "/kim"},
		{"/kim", "/kim"},
		{" /api  ", "/api"},
	}
	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.expected {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
