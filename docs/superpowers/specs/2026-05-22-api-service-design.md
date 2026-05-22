# API Service Design

## Overview

为 kim 项目搭建 HTTP API 服务框架，包括路由注册、中间件装配、统一响应格式，以及"发送私聊消息"接口（仅占位，不含业务逻辑）。

## Package Structure

```
internal/
├── handler/              # HTTP handler 层
│   ├── handler.go        # 共享依赖（logger 等）
│   └── message.go        # 消息相关 handler
├── router/
│   └── router.go         # 集中式路由注册
├── middleware/
│   └── middleware.go     # 中间件装配
└── model/
    └── response.go       # 统一请求/响应类型
```

## Route Design

基于已有 `route_prefix: "/kim"`，路由注册在 `router.go` 中集中管理：

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| `POST` | `/kim/v1/messages` | `message.Send` | 占位 |

## Middleware Chain

| Middleware | Source | Notes |
|------------|--------|-------|
| Recovery | Hertz built-in `recovery.Recovery()` | panic 恢复 |
| Logging | Custom (based on `slog`) | 记录 method/path/latency/status |
| Rate Limiter | `hertz-contrib/limiter` | 令牌桶，按 IP 限流 |
| CORS | `hertz-contrib/cors` | 开发阶段宽松配置 |

## Response Format

Unified envelope: `{"code": 0, "message": "ok", "data": {}}`

Error codes: `0` success, `400` bad request, `429` rate limited, `500` internal error.

## Wire Integration

Add providers to `wire.go` for handler, middleware, and router registration, injected into `App` initialization flow.

## POST /kim/v1/messages

Request:
```json
{
  "conversation_id": "conv_xxx",
  "receiver_id": "user_yyy",
  "content": "hello",
  "type": "text"
}
```

Response:
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

Handler does parameter binding and returns placeholder response only.
