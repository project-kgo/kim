package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kanengo/ku/mqx"
	"github.com/project-kgo/kim/internal/data"
	"github.com/project-kgo/kim/internal/event"
	"github.com/project-kgo/kim/internal/model"
)

// Consumer 消息消费者，统一注册项目相关的 topic 回调
type Consumer struct {
	logger        *slog.Logger
	messageStore  *data.MessageStore
	pubsub        mqx.PubSub
	subscriptions []mqx.Subscription
}

// NewConsumer 创建 Consumer 实例
func NewConsumer(logger *slog.Logger, messageStore *data.MessageStore, pubsub mqx.PubSub) *Consumer {
	return &Consumer{
		logger:       logger,
		messageStore: messageStore,
		pubsub:       pubsub,
	}
}

// Register 注册所有 topic 的回调处理
func (c *Consumer) Register(ctx context.Context) error {
	sub, err := c.pubsub.Subscribe(ctx, &mqx.SubscribeRequest{
		Topic:         event.TopicC2CMessage,
		ConsumerGroup: "kim-consumer",
		Handler:       c.handleC2CMessage,
	})
	if err != nil {
		return err
	}
	c.subscriptions = append(c.subscriptions, sub)

	c.logger.Info("consumer registered",
		slog.String("topic", event.TopicC2CMessage),
	)
	return nil
}

// handleC2CMessage 处理 C2C 消息：解析并存储到数据库
func (c *Consumer) handleC2CMessage(ctx context.Context, msg *mqx.Message) error {
	var evt event.MessageEvent
	if err := sonic.Unmarshal(msg.Data, &evt); err != nil {
		c.logger.ErrorContext(ctx, "failed to unmarshal message event",
			slog.String("error", err.Error()),
		)
		return err
	}

	createdAt := time.UnixMilli(evt.CreatedAt)

	senderID, _ := strconv.ParseInt(evt.SenderID, 10, 64)
	receiverID, _ := strconv.ParseInt(evt.ReceiverID, 10, 64)

	dbMsg := &model.Message{
		ID:             evt.MessageID,
		CreatedAt:      createdAt,
		ConversationID: evt.ConversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        evt.Content,
		Status:         1,
		UpdatedAt:      createdAt,
	}

	if err := c.messageStore.SaveMessage(ctx, dbMsg); err != nil {
		c.logger.ErrorContext(ctx, "failed to save message",
			slog.Int64("message_id", evt.MessageID),
			slog.String("error", err.Error()),
		)
		return err
	}

	c.logger.InfoContext(ctx, "message stored",
		slog.Int64("message_id", evt.MessageID),
		slog.String("conversation_id", evt.ConversationID),
	)

	return nil
}

// Close 关闭所有订阅
func (c *Consumer) Close() error {
	for _, sub := range c.subscriptions {
		if err := sub.Close(); err != nil {
			c.logger.Warn("failed to close subscription", slog.String("error", err.Error()))
		}
	}
	return nil
}
