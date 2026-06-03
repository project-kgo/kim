package service

import (
	"context"
	"errors"
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

type C2CMessageRecord = model.C2CMessageRecord

type c2cMessageStore interface {
	SaveC2CMessage(ctx context.Context, record model.C2CMessageRecord) error
}

type c2cMessagePusher interface {
	PushC2CMessage(ctx context.Context, evt event.MessageEvent) error
}

// Consumer 消息消费者，统一注册项目相关的 topic 回调
type Consumer struct {
	logger        *slog.Logger
	messageStore  c2cMessageStore
	pubsub        mqx.PubSub
	messagePusher c2cMessagePusher
	snowflakeNode *snowflakex.Node
	subscriptions []mqx.Subscription
}

// NewConsumer 创建 Consumer 实例
func NewConsumer(logger *slog.Logger, messageStore c2cMessageStore, pubsub mqx.PubSub, messagePusher c2cMessagePusher, snowflakeNode *snowflakex.Node) *Consumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Consumer{
		logger:        logger,
		messageStore:  messageStore,
		pubsub:        pubsub,
		messagePusher: messagePusher,
		snowflakeNode: snowflakeNode,
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

// handleC2CMessage 处理 C2C 消息：落库、同步邮箱、会话列表和接收者推送
func (c *Consumer) handleC2CMessage(ctx context.Context, msg *mqx.Message) error {
	var evt event.MessageEvent
	if err := sonic.Unmarshal(msg.Data, &evt); err != nil {
		c.logger.ErrorContext(ctx, "failed to unmarshal message event",
			slog.String("error", err.Error()),
		)
		return err
	}

	senderID, err := strconv.ParseInt(evt.SenderID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse sender_id: %w", err)
	}
	receiverID, err := strconv.ParseInt(evt.ReceiverID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse receiver_id: %w", err)
	}
	conversationID, err := C2CConversationID(senderID, receiverID)
	if err != nil {
		return err
	}
	evt.ConversationID = conversationID

	jsonContent, err := jsonString(evt.Content)
	if err != nil {
		return err
	}
	if c.snowflakeNode == nil {
		return errors.New("snowflake node is required")
	}
	if evt.MessageID <= 0 {
		return errors.New("message_id must be positive")
	}

	messageCreatedAt := c.snowflakeNode.GetTime(evt.MessageID)
	evt.CreatedAt = messageCreatedAt.UnixMilli()

	dbMsg := model.Message{
		ID:             evt.MessageID,
		CreatedAt:      messageCreatedAt,
		ConversationID: conversationID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        jsonContent,
		Status:         1,
		UpdatedAt:      messageCreatedAt,
	}

	record := model.C2CMessageRecord{
		Message: dbMsg,
		SyncMails: []model.UserSyncMail{
			newMessageSyncMail(c.snowflakeNode, senderID, senderID, conversationID, evt.MessageID, evt.Content),
			newMessageSyncMail(c.snowflakeNode, receiverID, senderID, conversationID, evt.MessageID, evt.Content),
		},
		Conversations: []model.Conversation{
			newConversation(senderID, receiverID, conversationID, evt.MessageID, evt.Content, 0, messageCreatedAt),
			newConversation(receiverID, senderID, conversationID, evt.MessageID, evt.Content, 1, messageCreatedAt),
		},
	}

	if err := c.messageStore.SaveC2CMessage(ctx, record); err != nil {
		c.logger.ErrorContext(ctx, "failed to save message",
			slog.Int64("message_id", evt.MessageID),
			slog.String("error", err.Error()),
		)
		return err
	}

	if c.messagePusher == nil {
		return errors.New("message pusher is required")
	}
	if err := c.messagePusher.PushC2CMessage(ctx, evt); err != nil {
		c.logger.ErrorContext(ctx, "failed to push c2c message",
			slog.Int64("message_id", evt.MessageID),
			slog.String("receiver_id", evt.ReceiverID),
			slog.String("error", err.Error()),
		)
		return err
	}

	c.logger.InfoContext(ctx, "message stored",
		slog.Int64("message_id", evt.MessageID),
		slog.String("conversation_id", conversationID),
	)

	return nil
}

func C2CConversationID(senderID, receiverID int64) (string, error) {
	if senderID <= 0 || receiverID <= 0 {
		return "", errors.New("sender_id and receiver_id must be positive")
	}
	if senderID > receiverID {
		senderID, receiverID = receiverID, senderID
	}
	return strconv.FormatInt(senderID, 10) + "-" + strconv.FormatInt(receiverID, 10), nil
}

func newMessageSyncMail(node *snowflakex.Node, userID, senderID int64, conversationID string, msgID int64, content string) model.UserSyncMail {
	seq := node.Generate()
	return model.UserSyncMail{
		SynSeq:         seq,
		UserID:         userID,
		CreatedAt:      node.GetTime(seq),
		SendID:         senderID,
		ConversationID: conversationID,
		SyncType:       1,
		MsgID:          msgID,
		Content:        content,
	}
}

func newConversation(userID, targetID int64, conversationID string, msgID int64, preview string, unread int, now time.Time) model.Conversation {
	return model.Conversation{
		UserID:         userID,
		ConversationID: conversationID,
		LastMsgID:      msgID,
		StartMsgID:     msgID,
		Preview:        preview,
		Unread:         unread,
		CreatedAt:      now,
		UpdatedAt:      now,
		TargetID:       targetID,
	}
}

func jsonString(content string) (string, error) {
	raw, err := sonic.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("encode message content: %w", err)
	}
	return string(raw), nil
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
