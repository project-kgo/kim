package kim

import (
	"testing"

	"github.com/project-kgo/kim/data"
)

func TestOptionsDefaults(t *testing.T) {
	opts, err := newOptions()
	if err == nil {
		t.Fatal("expected error")
	}
	_ = opts
}

func TestOptionsOverrideDefaults(t *testing.T) {
	opts, err := newOptions(
		WithRoutePrefix(" /chat "),
		WithGatewaySocket(" /tmp/custom-kim-gate.sock "),
		WithRedisDSN(" redis://localhost:6379/0 "),
		WithDBDSN(" postgres://kim:secret@localhost:5432/kim?sslmode=disable "),
	)
	if err != nil {
		t.Fatalf("newOptions returned error: %v", err)
	}

	if opts.RoutePrefix != "/chat" {
		t.Fatalf("RoutePrefix = %q, want %q", opts.RoutePrefix, "/chat")
	}
	if opts.GatewaySocket != "/tmp/custom-kim-gate.sock" {
		t.Fatalf("GatewaySocket = %q, want %q", opts.GatewaySocket, "/tmp/custom-kim-gate.sock")
	}
	if opts.RedisDSN != "redis://localhost:6379/0" {
		t.Fatalf("RedisDSN = %q, want %q", opts.RedisDSN, "redis://localhost:6379/0")
	}
	if opts.DBDSN != "postgres://kim:secret@localhost:5432/kim?sslmode=disable" {
		t.Fatalf("DBDSN = %q, want %q", opts.DBDSN, "postgres://kim:secret@localhost:5432/kim?sslmode=disable")
	}
}

func TestOptionsAcceptsDataClientWithoutDSNs(t *testing.T) {
	opts, err := newOptions(WithDataClient(&data.Client{}))
	if err != nil {
		t.Fatalf("newOptions returned error: %v", err)
	}

	if opts.DataClient == nil {
		t.Fatal("DataClient is nil")
	}
}

func TestOptionsValidation(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "empty route prefix",
			opts: []Option{WithRoutePrefix(" ")},
		},
		{
			name: "route prefix without slash",
			opts: []Option{WithRoutePrefix("kim")},
		},
		{
			name: "empty gateway socket",
			opts: []Option{WithGatewaySocket(" ")},
		},
		{
			name: "empty redis dsn",
			opts: []Option{
				WithRedisDSN(" "),
				WithDBDSN("postgres://kim:secret@localhost:5432/kim?sslmode=disable"),
			},
		},
		{
			name: "empty db dsn",
			opts: []Option{
				WithRedisDSN("redis://localhost:6379/0"),
				WithDBDSN(" "),
			},
		},
		{
			name: "nil option",
			opts: []Option{nil},
		},
		{
			name: "nil data client",
			opts: []Option{WithDataClient(nil)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := newOptions(tt.opts...); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
