package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/project-kgo/kim/internal/config"
	"github.com/project-kgo/kim/internal/data"
	"github.com/project-kgo/kim/internal/gateway"
)

func TestNewApp(t *testing.T) {
	app := New(config.Defaults(), slog.Default(), &data.Data{}, &gateway.Client{}, nil, nil)
	if app == nil {
		t.Fatal("New returned nil")
	}
}

func TestStartShutdown(t *testing.T) {
	app := New(config.Defaults(), slog.Default(), &data.Data{}, &gateway.Client{}, nil, nil)
	if err := app.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}

func TestShutdownIdempotent(t *testing.T) {
	app := New(config.Defaults(), slog.Default(), &data.Data{}, &gateway.Client{}, nil, nil)
	_ = app.Start()
	time.Sleep(50 * time.Millisecond)
	ctx := context.Background()
	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("first Shutdown returned error: %v", err)
	}
	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown returned error: %v", err)
	}
}

func TestNilAppSafe(t *testing.T) {
	var app *App
	if err := app.Start(); err == nil {
		t.Fatal("expected error from nil app")
	}
	if app.Done() != nil {
		t.Fatal("Done should return nil for nil app")
	}
	if err := app.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown on nil app returned error: %v", err)
	}
}
