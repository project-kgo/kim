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
