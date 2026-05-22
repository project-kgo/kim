# API Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 HTTP API 服务框架：统一响应格式、四个中间件（Recovery/Logging/限流/CORS）、路由集中注册、发送私聊消息占位接口。

**Architecture:** 新建 `internal/model/`、`internal/middleware/`、`internal/handler/`、`internal/router/` 四个包。`handler` 和路由注册在 `app.New()` 内部完成，不改变现有 Wire 图。中间件按 Recovery → Logging → CORS → RateLimiter 顺序注册，确保 panic 在最外层捕获、日志记录所有请求、CORS 先于限流。

**Tech Stack:** Go 1.26, CloudWeGo Hertz v0.10.4, Google Wire v0.7.0, log/slog

---

### Task 1: 创建统一响应类型 `internal/model/response.go`

**Files:**
- Create: `internal/model/response.go`

- [ ] **Step 1: 创建文件**

```go
package model

import (
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

// 预定义错误码
const (
	CodeSuccess       = 0
	CodeBadRequest    = 400
	CodeRateLimited   = 429
	CodeInternalError = 500
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Success 成功响应
func Success(ctx *app.RequestContext, data any) {
	ctx.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "ok",
		Data:    data,
	})
}

// Error 错误响应
func Error(ctx *app.RequestContext, code int, msg string) {
	httpStatus := http.StatusOK
	if code == CodeInternalError {
		httpStatus = http.StatusInternalServerError
	}
	ctx.JSON(httpStatus, Response{
		Code:    code,
		Message: msg,
	})
}

// Abort 中断并返回错误响应
func Abort(ctx *app.RequestContext, code int, msg string) {
	ctx.Abort()
	Error(ctx, code, msg)
}
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/model/`

---

### Task 2: 创建中间件 `internal/middleware/middleware.go`

**Files:**
- Create: `internal/middleware/middleware.go`

- [ ] **Step 1: 创建中间件包**

```go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
)

// Logging 基于 slog 的请求日志中间件
func Logging(logger *slog.Logger) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		start := time.Now()
		ctx.Next(c)
		logger.Info("request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.Int("status", ctx.Response.StatusCode()),
			slog.Duration("latency", time.Since(start)),
		)
	}
}

// CORS 跨域中间件，开发阶段允许所有来源
func CORS() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if string(ctx.Method()) == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next(c)
	}
}

// visitor 记录每个 IP 的令牌状态
type visitor struct {
	lastSeen time.Time
	tokens   float64
}

// RateLimiter 基于令牌桶的 IP 限流中间件
// rate: 每秒填充令牌数, burst: 桶容量
func RateLimiter(rate int, burst int) app.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*visitor)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c context.Context, ctx *app.RequestContext) {
		ip := ctx.ClientIP()
		mu.Lock()
		v, exists := visitors[ip]
		now := time.Now()
		if !exists {
			v = &visitor{lastSeen: now, tokens: float64(burst) - 1}
			visitors[ip] = v
			mu.Unlock()
			ctx.Next(c)
			return
		}
		elapsed := now.Sub(v.lastSeen).Seconds()
		v.tokens = min(float64(burst), v.tokens+elapsed*float64(rate)) - 1
		v.lastSeen = now
		if v.tokens < 0 {
			mu.Unlock()
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]any{
				"code":    429,
				"message": "rate limit exceeded",
			})
			return
		}
		mu.Unlock()
		ctx.Next(c)
	}
}

// Recovery 返回 Hertz 内置的 panic 恢复中间件
func Recovery() app.HandlerFunc {
	return recovery.Recovery()
}
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/middleware/`

---

### Task 3: 创建 Handler 基础结构 `internal/handler/handler.go`

**Files:**
- Create: `internal/handler/handler.go`

- [ ] **Step 1: 创建 Handler 结构体**

```go
package handler

import "log/slog"

// Handler 持有所有 handler 共享的依赖
type Handler struct {
	logger *slog.Logger
}

// New 创建 Handler 实例
func New(logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{logger: logger}
}
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/handler/`

---

### Task 4: 创建发送消息 Handler `internal/handler/message.go`

**Files:**
- Create: `internal/handler/message.go`

- [ ] **Step 1: 创建消息 handler（占位逻辑）**

```go
package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/project-kgo/kim/internal/model"
)

// SendMessageRequest 发送消息请求体
type SendMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	ReceiverID     string `json:"receiver_id"`
	Content        string `json:"content"`
	Type           string `json:"type"`
}

// SendMessageResponse 发送消息响应体
type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	CreatedAt string `json:"created_at"`
}

// SendMessage 发送私聊消息（占位）
func (h *Handler) SendMessage(ctx context.Context, c *app.RequestContext) {
	var req SendMessageRequest
	if err := c.BindJSON(&req); err != nil {
		model.Error(c, model.CodeBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.ConversationID == "" {
		model.Error(c, model.CodeBadRequest, "conversation_id is required")
		return
	}
	if req.ReceiverID == "" {
		model.Error(c, model.CodeBadRequest, "receiver_id is required")
		return
	}

	h.logger.InfoContext(ctx, "send message handler invoked",
		slog.String("conversation_id", req.ConversationID),
		slog.String("receiver_id", req.ReceiverID),
		slog.String("type", req.Type),
	)

	model.Success(c, SendMessageResponse{})
}
```

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/handler/`

---

### Task 5: 创建路由注册 `internal/router/router.go`

**Files:**
- Create: `internal/router/router.go`

- [ ] **Step 1: 创建路由注册函数**

```go
package router

import (
	"log/slog"

	"github.com/cloudwego/hertz/pkg/route"
	"github.com/project-kgo/kim/internal/handler"
	"github.com/project-kgo/kim/internal/middleware"
)

const (
	defaultRateLimit = 100  // 每秒令牌数
	defaultBurst     = 200  // 桶容量
)

// Register 注册所有路由到 Hertz 引擎
func Register(routerGroup route.IRouter, h *handler.Handler, logger *slog.Logger) {
	// 注册全局中间件
	routerGroup.Use(
		middleware.Recovery(),
		middleware.Logging(logger),
		middleware.CORS(),
		middleware.RateLimiter(defaultRateLimit, defaultBurst),
	)

	// v1 API 路由组
	v1 := routerGroup.Group("/kim/v1")
	{
		v1.POST("/messages", h.SendMessage)
	}
}
```

注意：Hertz 的 `*server.Hertz` 实现了 `IRouter` 接口（因为它嵌入了 `*route.Engine`），而 `IRouter` 包含 `IRoutes` 和 `Group()`。

- [ ] **Step 2: 验证编译**

Run: `go build ./internal/router/`

---

### Task 6: 修改 `internal/app/app.go` 集成路由注册

**Files:**
- Modify: `internal/app/app.go`

- [ ] **Step 1: 在 New() 中注册路由**

将原来的 `internal/app/app.go` 做以下修改：

**修改前：**
```go
import (
    "context"
    "errors"
    "log/slog"
    "sync"

    hertzserver "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/project-kgo/kim/internal/config"
    "github.com/project-kgo/kim/internal/data"
    "github.com/project-kgo/kim/internal/gateway"
)
```

**修改后——增加三行 import：**
```go
import (
    "context"
    "errors"
    "log/slog"
    "sync"

    hertzserver "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/project-kgo/kim/internal/config"
    "github.com/project-kgo/kim/internal/data"
    "github.com/project-kgo/kim/internal/gateway"
    "github.com/project-kgo/kim/internal/handler"
    "github.com/project-kgo/kim/internal/router"
)
```

**修改 New() 函数，在创建 http server 后、返回 App 前插入路由注册：**

在 `app.go` 中找到：
```go
    http := hertzserver.New(hertzserver.WithHostPorts(cfg.HTTPAddr))
    return &App{
```

替换为：
```go
    http := hertzserver.New(hertzserver.WithHostPorts(cfg.HTTPAddr))
    h := handler.New(logger)
    router.Register(http, h, logger)
    return &App{
```

- [ ] **Step 2: 验证编译**

Run: `go build ./...`

---

### Task 7: 生成 Wire 代码并验证构建

**Files:**
- No changes to wire.go (graph unchanged since handler/router are created inside app.New())

- [ ] **Step 1: 重新生成 wire_gen.go**

Run: `go generate ./...`

Expected: wire 生成成功，wire_gen.go 内容不变（因为 wire.go 的依赖图未变）

- [ ] **Step 2: 完整构建验证**

Run: `go build -o /dev/null ./main.go`

Expected: 编译成功，无错误

- [ ] **Step 3: 完整测试运行**

Run: `go test ./...`

Expected: 所有已有测试通过

---

### Task 8: 手动验证 API 端点

- [ ] **Step 1: 启动服务**

Run: `go run . &`
Wait: 看到 "hertz server started" 日志

- [ ] **Step 2: 测试 POST /kim/v1/messages**

Run:
```bash
curl -s -X POST http://localhost:8080/kim/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"conversation_id":"conv_123","receiver_id":"user_456","content":"hello","type":"text"}' | python3 -m json.tool
```

Expected:
```json
{
    "code": 0,
    "message": "ok",
    "data": {
        "message_id": "",
        "created_at": ""
    }
}
```

- [ ] **Step 3: 测试参数校验**

Run:
```bash
curl -s -X POST http://localhost:8080/kim/v1/messages \
  -H "Content-Type: application/json" \
  -d '{}' | python3 -m json.tool
```

Expected:
```json
{
    "code": 400,
    "message": "conversation_id is required"
}
```

- [ ] **Step 4: 测试 CORS 预检请求**

Run:
```bash
curl -s -X OPTIONS http://localhost:8080/kim/v1/messages -i | head -10
```

Expected: 返回 204，响应头包含 CORS 头

- [ ] **Step 5: 停止服务**

Run: `kill %1`
