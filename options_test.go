package kim

import "testing"

func TestOptionsDefaults(t *testing.T) {
	opts, err := newOptions()
	if err != nil {
		t.Fatalf("newOptions returned error: %v", err)
	}

	if opts.RoutePrefix != "/kim" {
		t.Fatalf("RoutePrefix = %q, want %q", opts.RoutePrefix, "/kim")
	}
	if opts.GatewaySocket != "/tmp/kim-gate.sock" {
		t.Fatalf("GatewaySocket = %q, want %q", opts.GatewaySocket, "/tmp/kim-gate.sock")
	}
}

func TestOptionsOverrideDefaults(t *testing.T) {
	opts, err := newOptions(
		WithRoutePrefix(" /chat "),
		WithGatewaySocket(" /tmp/custom-kim-gate.sock "),
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
			name: "nil option",
			opts: []Option{nil},
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
