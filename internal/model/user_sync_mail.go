package model

import "time"

// UserSyncMail 对应 dim.user_sync_mail 表
type UserSyncMail struct {
	SynSeq         int64     `json:"syn_seq"`
	UserID         int64     `json:"user_id"`
	CreatedAt      time.Time `json:"created_at"`
	SendID         int64     `json:"send_id"`
	ConversationID string    `json:"conversation_id"`
	SyncType       int       `json:"sync_type"` //1-新增消息 2-删除消息 3-更新消息 4-撤回消息
	MsgID          int64     `json:"msg_id"`
	Content        string    `json:"content"`
}
