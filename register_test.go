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

	if err := Register(h); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
}

func TestRegisterAppliesOptions(t *testing.T) {
	h := server.New()

	err := Register(
		h,
		WithRoutePrefix("/chat"),
		WithGatewaySocket("/tmp/custom-kim-gate.sock"),
	)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
}
