package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kanengo/ku/mqx"
	"github.com/kanengo/ku/snowflakex"
	"github.com/project-kgo/kim/internal/event"
	"github.com/project-kgo/kim/internal/model"
)

// MessageService 消息业务逻辑
type MessageService struct {
	logger        *slog.Logger
	snowflakeNode *snowflakex.Node
	pubsub        mqx.PubSub
}

// NewMessageService 创建 MessageService 实例
func NewMessageService(logger *slog.Logger, snowflakeNode *snowflakex.Node, pubsub mqx.PubSub) *MessageService {
	return &MessageService{
		logger:        logger,
		snowflakeNode: snowflakeNode,
		pubsub:        pubsub,
	}
}

// Send 发送私聊消息
func (s *MessageService) Send(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error) {
	msgID := s.snowflakeNode.Generate()
	now := time.Now()
	msgIDStr := strconv.FormatInt(msgID, 10)

	payload := event.MessageEvent{
		MessageID:      msgID,
		ConversationID: req.ConversationID,
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		Content:        req.Content,
		Type:           req.Type,
		CreatedAt:      now.UnixMilli(),
	}

	data, err := sonic.Marshal(payload)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal message payload",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("encode message: %w", err)
	}

	if err := s.pubsub.Publish(ctx, &mqx.PublishRequest{
		ID:          msgIDStr,
		Topic:       event.TopicC2CMessage,
		Data:        data,
		ContentType: "application/json",
	}); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish message",
			slog.String("message_id", msgIDStr),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("publish message: %w", err)
	}

	s.logger.InfoContext(ctx, "message sent",
		slog.String("message_id", msgIDStr),
		slog.String("conversation_id", req.ConversationID),
		slog.String("sender_id", req.SenderID),
		slog.String("receiver_id", req.ReceiverID),
	)

	return &model.SendMessageResponse{
		MessageID: msgIDStr,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}
