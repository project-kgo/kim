package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kanengo/ku/mqx"
	"github.com/kanengo/ku/snowflakex"
	kimgatev1 "github.com/project-kgo/kim-gate/proto/kimgate/v1"
	"github.com/project-kgo/kim/internal/data"
	"github.com/project-kgo/kim/internal/event"
	"github.com/project-kgo/kim/internal/model"
	kimv1 "github.com/project-kgo/kim/proto/kim/v1"
	"google.golang.org/protobuf/proto"
)

const C2CMessagePushMethod = "message.c2c.new"
const c2cDedupTTL = 5 * time.Minute

var ErrClientMessageConflict = errors.New("client_msg_id payload conflict")

type gatewayPusher interface {
	SendToUsers(ctx context.Context, req *kimgatev1.SendToUsersRequest) (*kimgatev1.SendResponse, error)
}

// MessagePusher 封装消息推送到 gateway 的协议细节。
type MessagePusher struct {
	gateway gatewayPusher
}

func NewMessagePusher(gateway gatewayPusher) *MessagePusher {
	return &MessagePusher{gateway: gateway}
}

func (p *MessagePusher) PushC2CMessage(ctx context.Context, evt event.MessageEvent) error {
	if p == nil || p.gateway == nil {
		return errors.New("gateway pusher is required")
	}
	payload, err := proto.Marshal(&kimv1.C2CMessagePush{
		MessageId:      evt.MessageID,
		ConversationId: evt.ConversationID,
		SenderId:       evt.SenderID,
		ReceiverId:     evt.ReceiverID,
		Content:        evt.Content,
		Type:           evt.Type,
		CreatedAt:      evt.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("encode gateway protobuf payload: %w", err)
	}
	_, err = p.gateway.SendToUsers(ctx, &kimgatev1.SendToUsersRequest{
		UserIds: []string{evt.ReceiverID},
		Method:  C2CMessagePushMethod,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("send gateway message: %w", err)
	}
	return nil
}

// MessageService 消息业务逻辑
type MessageService struct {
	logger        *slog.Logger
	snowflakeNode *snowflakex.Node
	pubsub        mqx.PubSub
	dedupStore    *data.MessageDedupStore
}

// NewMessageService 创建 MessageService 实例
func NewMessageService(logger *slog.Logger, snowflakeNode *snowflakex.Node, pubsub mqx.PubSub, dedupStore *data.MessageDedupStore) *MessageService {
	return &MessageService{
		logger:        logger,
		snowflakeNode: snowflakeNode,
		pubsub:        pubsub,
		dedupStore:    dedupStore,
	}
}

// Send 发送私聊消息
func (s *MessageService) Send(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error) {
	msgID := s.snowflakeNode.Generate()
	now := time.Now()
	msgIDStr := strconv.FormatInt(msgID, 10)
	senderID, err := strconv.ParseInt(req.SenderID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse sender_id: %w", err)
	}
	receiverID, err := strconv.ParseInt(req.ReceiverID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse receiver_id: %w", err)
	}
	conversationID, err := C2CConversationID(senderID, receiverID)
	if err != nil {
		return nil, err
	}

	payload := event.MessageEvent{
		MessageID:      msgID,
		ConversationID: conversationID,
		ClientMsgID:    req.ClientMsgID,
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		Content:        req.Content,
		Type:           req.MessageType,
		CreatedAt:      now.UnixMilli(),
	}

	fingerprint := c2cMessageFingerprint(req)
	if s.dedupStore != nil {
		ok, err := s.dedupStore.ReserveC2CMessage(ctx, req.SenderID, req.ClientMsgID, data.MessageDedupValue{
			MessageID:   msgIDStr,
			CreatedAt:   now.Format(time.RFC3339),
			Fingerprint: fingerprint,
		}, c2cDedupTTL)
		if err != nil {
			return nil, err
		}
		if !ok {
			return s.handleDuplicateC2CMessage(ctx, req, fingerprint)
		}
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
		if s.dedupStore != nil {
			if delErr := s.dedupStore.DeleteC2CMessage(ctx, req.SenderID, req.ClientMsgID); delErr != nil {
				s.logger.WarnContext(ctx, "failed to delete c2c dedup key after publish failure",
					slog.String("sender_id", req.SenderID),
					slog.String("client_msg_id", req.ClientMsgID),
					slog.String("error", delErr.Error()),
				)
			}
		}
		s.logger.ErrorContext(ctx, "failed to publish message",
			slog.String("message_id", msgIDStr),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("publish message: %w", err)
	}

	s.logger.InfoContext(ctx, "message sent",
		slog.String("message_id", msgIDStr),
		slog.String("conversation_id", conversationID),
		slog.String("sender_id", req.SenderID),
		slog.String("receiver_id", req.ReceiverID),
	)

	return &model.SendMessageResponse{
		MessageID: msgIDStr,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}

func (s *MessageService) handleDuplicateC2CMessage(ctx context.Context, req model.SendMessageRequest, fingerprint string) (*model.SendMessageResponse, error) {
	value, err := s.dedupStore.GetC2CMessage(ctx, req.SenderID, req.ClientMsgID)
	if err != nil {
		return nil, err
	}
	if value.Fingerprint != fingerprint {
		return nil, ErrClientMessageConflict
	}
	return &model.SendMessageResponse{
		MessageID: value.MessageID,
		CreatedAt: value.CreatedAt,
	}, nil
}

func c2cMessageFingerprint(req model.SendMessageRequest) string {
	h := sha256.New()
	for _, part := range []string{req.SenderID, req.ReceiverID, req.Content, req.MessageType} {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
