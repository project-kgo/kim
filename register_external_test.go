package kim_test

import (
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/project-kgo/kim"
	"github.com/project-kgo/kim/data"
)

func TestRegisterAcceptsExternalDataClient(t *testing.T) {
	h := server.New()

	err := kim.Register(
		h,
		kim.WithDataClient(&data.Client{}),
	)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if len(h.Engine.OnShutdown) != 0 {
		t.Fatalf("OnShutdown hooks = %d, want %d", len(h.Engine.OnShutdown), 0)
	}
}
