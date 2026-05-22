package router

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/ut"
	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/project-kgo/kim/internal/handler"
	"github.com/project-kgo/kim/internal/model"
)

func newTestServer() *hertzserver.Hertz {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	h := handler.New(logger)
	srv := hertzserver.New()
	Register(srv, h, logger, "/kim")
	return srv
}

func TestSendMessageSuccess(t *testing.T) {
	srv := newTestServer()
	body := `{"conversation_id":"conv_123","receiver_id":"user_456","content":"hello","type":"text"}`
	rec := ut.PerformRequest(srv.Engine, "POST", "/kim/v1/messages",
		&ut.Body{Body: strings.NewReader(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"},
	)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Result().Body())
	}
	var resp model.Response
	if err := json.Unmarshal(rec.Result().Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != model.CodeSuccess {
		t.Fatalf("expected code 0, got %d: %+v", resp.Code, resp)
	}
	t.Logf("success response: %+v", resp)
}

func TestSendMessageValidation(t *testing.T) {
	srv := newTestServer()
	body := `{}`
	rec := ut.PerformRequest(srv.Engine, "POST", "/kim/v1/messages",
		&ut.Body{Body: strings.NewReader(body), Len: len(body)},
		ut.Header{Key: "Content-Type", Value: "application/json"},
	)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp model.Response
	if err := json.Unmarshal(rec.Result().Body(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Code != model.CodeBadRequest {
		t.Fatalf("expected code %d, got %d: %+v", model.CodeBadRequest, resp.Code, resp)
	}
	t.Logf("validation response: %+v", resp)
}

func TestCORSPreflight(t *testing.T) {
	srv := newTestServer()
	rec := ut.PerformRequest(srv.Engine, "OPTIONS", "/kim/v1/messages", nil)
	if rec.Code != 204 {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected CORS origin '*', got '%s'", origin)
	}
	t.Logf("CORS preflight OK: origin=%s", origin)
}
