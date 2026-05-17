package gateway

import "testing"

func TestUnixTarget(t *testing.T) {
	target := UnixTarget("/tmp/kim-gate.sock")
	if target != "unix:///tmp/kim-gate.sock" {
		t.Fatalf("UnixTarget = %q, want %q", target, "unix:///tmp/kim-gate.sock")
	}
}

func TestUnixTargetTrimsSpace(t *testing.T) {
	target := UnixTarget(" /tmp/kim-gate.sock ")
	if target != "unix:///tmp/kim-gate.sock" {
		t.Fatalf("UnixTarget = %q, want %q", target, "unix:///tmp/kim-gate.sock")
	}
}

func TestNewClientRequiresSocket(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected error")
	}
}
