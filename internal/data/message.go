package data

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/project-kgo/kim/internal/model"
)

// MessageStore 消息持久化
type MessageStore struct {
	db *sqlx.DB
}

// NewMessageStore 创建 MessageStore 实例
func NewMessageStore(db *sqlx.DB) *MessageStore {
	return &MessageStore{db: db}
}

// SaveMessage 保存消息到数据库
func (s *MessageStore) SaveMessage(ctx context.Context, msg *model.Message) error {
	_, err := s.db.NamedExecContext(ctx,
		`INSERT INTO dim.messages (id, created_at, conversation_id, sender_id, receiver_id, content, status, updated_at)
		 VALUES (:id, :created_at, :conversation_id, :sender_id, :receiver_id, :content, :status, :updated_at)`,
		msg,
	)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}
