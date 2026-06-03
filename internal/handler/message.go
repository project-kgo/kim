package handler

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/project-kgo/kim/internal/model"
	"github.com/project-kgo/kim/internal/service"
)

// SendMessage 发送私聊消息
func (h *Handler) SendMessage(ctx context.Context, c *app.RequestContext) {
	var req model.SendMessageRequest
	if err := c.BindJSON(&req); err != nil {
		model.Error(c, model.CodeBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.SenderID == "" {
		model.Error(c, model.CodeBadRequest, "sender_id is required")
		return
	}
	if req.ReceiverID == "" {
		model.Error(c, model.CodeBadRequest, "receiver_id is required")
		return
	}
	if req.ClientMsgID == "" {
		model.Error(c, model.CodeBadRequest, "client_msg_id is required")
		return
	}

	resp, err := h.messageService.Send(ctx, req)
	if err != nil {
		if errors.Is(err, service.ErrClientMessageConflict) {
			model.Error(c, model.CodeBadRequest, "client_msg_id payload conflict")
			return
		}
		model.Error(c, model.CodeInternalError, "failed to send message")
		return
	}

	model.Success(c, resp)
}
