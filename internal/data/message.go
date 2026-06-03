package data

import (
	"context"
	"fmt"
	"hash/fnv"

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
		 VALUES (:id, :created_at, :conversation_id, :sender_id, :receiver_id, CAST(:content AS jsonb), :status, :updated_at)`,
		msg,
	)
	if err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	return nil
}

// SaveC2CMessage 幂等保存私聊消息、同步邮箱和双方会话列表。
func (s *MessageStore) SaveC2CMessage(ctx context.Context, record model.C2CMessageRecord) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin c2c message transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, record.Message.ID); err != nil {
		return fmt.Errorf("lock c2c message: %w", err)
	}

	inserted, err := insertMessageIfAbsent(ctx, tx, record.Message)
	if err != nil {
		return err
	}
	if !inserted {
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit duplicate c2c message transaction: %w", err)
		}
		return nil
	}
	for _, mail := range record.SyncMails {
		if err = insertSyncMailIfAbsent(ctx, tx, mail); err != nil {
			return err
		}
	}
	for _, conversation := range record.Conversations {
		if err = upsertConversation(ctx, tx, conversation); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit c2c message transaction: %w", err)
	}
	return nil
}

func insertMessageIfAbsent(ctx context.Context, tx *sqlx.Tx, msg model.Message) (bool, error) {
	result, err := tx.NamedExecContext(ctx,
		`INSERT INTO dim.messages (id, created_at, conversation_id, sender_id, receiver_id, content, status, updated_at)
		 SELECT :id, :created_at, :conversation_id, :sender_id, :receiver_id, CAST(:content AS jsonb), :status, :updated_at
		 WHERE NOT EXISTS (SELECT 1 FROM dim.messages WHERE id = :id)`,
		msg,
	)
	if err != nil {
		return false, fmt.Errorf("insert c2c message: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read message insert result: %w", err)
	}
	return affected > 0, nil
}

func insertSyncMailIfAbsent(ctx context.Context, tx *sqlx.Tx, mail model.UserSyncMail) error {
	_, err := tx.NamedExecContext(ctx,
		`INSERT INTO dim.user_sync_mail (syn_seq, user_id, created_at, send_id, conversation_id, sync_type, msg_id, content)
		 SELECT :syn_seq, :user_id, :created_at, :send_id, :conversation_id, :sync_type, :msg_id, :content
		 WHERE NOT EXISTS (
		 	SELECT 1 FROM dim.user_sync_mail
		 	WHERE user_id = :user_id AND msg_id = :msg_id AND sync_type = :sync_type
		 )`,
		mail,
	)
	if err != nil {
		return fmt.Errorf("insert user sync mail: %w", err)
	}
	return nil
}

func upsertConversation(ctx context.Context, tx *sqlx.Tx, conversation model.Conversation) error {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, conversationLockKey(conversation.UserID, conversation.ConversationID)); err != nil {
		return fmt.Errorf("lock conversation: %w", err)
	}

	result, err := tx.NamedExecContext(ctx,
		`UPDATE dim.conversations
		 SET last_msg_id = CASE
		     	WHEN :last_msg_id > last_msg_id THEN :last_msg_id
		     	ELSE last_msg_id
		     END,
		     preview = CASE
		     	WHEN :last_msg_id > last_msg_id THEN :preview
		     	ELSE preview
		     END,
		     unread = unread + :unread,
		     updated_at = CASE
		     	WHEN :last_msg_id > last_msg_id THEN :updated_at
		     	ELSE updated_at
		     END,
		     target_id = :target_id
		 WHERE user_id = :user_id AND conversation_id = :conversation_id`,
		conversation,
	)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read conversation update result: %w", err)
	}
	if affected > 0 {
		return nil
	}

	_, err = tx.NamedExecContext(ctx,
		`INSERT INTO dim.conversations
		 (user_id, conversation_id, last_msg_id, start_msg_id, preview, unread, pinnd_at, created_at, updated_at, target_id)
		 SELECT :user_id, :conversation_id, :last_msg_id, :start_msg_id, :preview, :unread, :pinnd_at, :created_at, :updated_at, :target_id
		 WHERE NOT EXISTS (
		 	SELECT 1 FROM dim.conversations
		 	WHERE user_id = :user_id AND conversation_id = :conversation_id
		 )`,
		conversation,
	)
	if err != nil {
		return fmt.Errorf("insert conversation: %w", err)
	}
	return nil
}

func conversationLockKey(userID int64, conversationID string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s", userID, conversationID)))
	return int64(h.Sum64())
}
