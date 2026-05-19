package kim

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func TestRegisterRequiresHertz(t *testing.T) {
	if err := Register(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestRegisterWithDefaults(t *testing.T) {
	h := server.New()

	if err := Register(h); err == nil {
		t.Fatal("expected error")
	}
}

func TestRegisterAppliesOptions(t *testing.T) {
	h := server.New()

	err := Register(
		h,
		WithRoutePrefix("/chat"),
		WithGatewaySocket("/tmp/custom-kim-gate.sock"),
		WithRedisDSN("redis://localhost:6379/0"),
		WithDBDSN("postgres://kim:secret@localhost:5432/kim?sslmode=disable"),
	)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
}

func TestRegisterAddsShutdownHook(t *testing.T) {
	h := server.New()

	err := Register(
		h,
		WithRedisDSN("redis://localhost:6379/0"),
		WithDBDSN("postgres://kim:secret@localhost:5432/kim?sslmode=disable"),
	)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if len(h.Engine.OnShutdown) != 1 {
		t.Fatalf("OnShutdown hooks = %d, want %d", len(h.Engine.OnShutdown), 1)
	}
}
