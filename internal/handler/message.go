package handler

import (
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/project-kgo/kim/internal/model"
)

// SendMessage 发送私聊消息（占位）
func (h *Handler) SendMessage(ctx context.Context, c *app.RequestContext) {
	var req model.SendMessageRequest
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

	model.Success(c, model.SendMessageResponse{})
}
