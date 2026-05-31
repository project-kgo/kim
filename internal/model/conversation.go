package model

import "time"

// Conversation 对应 dim.conversations 表
type Conversation struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	ConversationID string     `json:"conversation_id"`
	LastMsgID      int64      `json:"last_msg_id"`
	StartMsgID     int64      `json:"start_msg_id"`
	Preview        string     `json:"preview"`
	Unread         int        `json:"unread"`
	PinnedAt       *time.Time `json:"pinnd_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	TargetID       int64      `json:"target_id"`
}
