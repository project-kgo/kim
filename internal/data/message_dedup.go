package data

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

// MessageDedupStore 负责消息短窗口去重的 Redis 读写。
type MessageDedupStore struct {
	redis messageDedupRedis
}

type MessageDedupValue struct {
	MessageID   string `json:"message_id"`
	CreatedAt   string `json:"created_at"`
	Fingerprint string `json:"fingerprint"`
}

type messageDedupRedis interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

func NewMessageDedupStore(client *redis.Client) *MessageDedupStore {
	return newMessageDedupStore(client)
}

func newMessageDedupStore(client messageDedupRedis) *MessageDedupStore {
	return &MessageDedupStore{redis: client}
}

func (s *MessageDedupStore) ReserveC2CMessage(ctx context.Context, senderID string, clientMsgID string, value MessageDedupValue, ttl time.Duration) (bool, error) {
	raw, err := sonic.MarshalString(value)
	if err != nil {
		return false, fmt.Errorf("encode c2c dedup value: %w", err)
	}
	ok, err := s.redis.SetNX(ctx, c2cMessageDedupKey(senderID, clientMsgID), raw, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("set c2c dedup key: %w", err)
	}
	return ok, nil
}

func (s *MessageDedupStore) GetC2CMessage(ctx context.Context, senderID string, clientMsgID string) (MessageDedupValue, error) {
	raw, err := s.redis.Get(ctx, c2cMessageDedupKey(senderID, clientMsgID)).Result()
	if err != nil {
		return MessageDedupValue{}, fmt.Errorf("get c2c dedup key: %w", err)
	}

	var value MessageDedupValue
	if err := sonic.UnmarshalString(raw, &value); err != nil {
		return MessageDedupValue{}, fmt.Errorf("decode c2c dedup value: %w", err)
	}
	return value, nil
}

func (s *MessageDedupStore) DeleteC2CMessage(ctx context.Context, senderID string, clientMsgID string) error {
	if _, err := s.redis.Del(ctx, c2cMessageDedupKey(senderID, clientMsgID)).Result(); err != nil {
		return fmt.Errorf("delete c2c dedup key: %w", err)
	}
	return nil
}

func c2cMessageDedupKey(senderID string, clientMsgID string) string {
	return "kim:c2c:dedup:" + senderID + ":" + clientMsgID
}
