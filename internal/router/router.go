package router

import (
	"log/slog"

	"github.com/cloudwego/hertz/pkg/route"
	"github.com/project-kgo/kim/internal/handler"
	"github.com/project-kgo/kim/internal/middleware"
)

const (
	defaultRateLimit = 100 // 每秒令牌数
	defaultBurst     = 200 // 桶容量
)

// Register 注册所有路由到 Hertz 引擎
func Register(routerGroup route.IRouter, h *handler.Handler, logger *slog.Logger, routePrefix string) {
	// 注册全局中间件
	routerGroup.Use(
		middleware.Recovery(),
		middleware.Logging(logger),
		middleware.CORS(),
		middleware.RateLimiter(defaultRateLimit, defaultBurst),
	)

	// v1 API 路由组
	v1 := routerGroup.Group(routePrefix + "/v1")
	{
		// c2c
		v1.POST("/c2c/messages", h.SendMessage)
	}
}
