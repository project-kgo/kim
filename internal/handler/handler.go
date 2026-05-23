package handler

import (
	"log/slog"

	"github.com/project-kgo/kim/internal/service"
)

// Handler 持有所有 handler 共享的依赖
type Handler struct {
	logger         *slog.Logger
	messageService *service.MessageService
}

// New 创建 Handler 实例
func New(logger *slog.Logger, messageService *service.MessageService) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		logger:         logger,
		messageService: messageService,
	}
}
