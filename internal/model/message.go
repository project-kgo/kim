package model

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id"`
	Content        string `json:"content"`
	Type           string `json:"type"`
}

// SendMessageResponse 发送消息响应体
type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	CreatedAt string `json:"created_at"`
}

// MessagePayload 发送到消息队列的消息体
type MessagePayload struct {
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id"`
	Content        string `json:"content"`
	Type           string `json:"type"`
	CreatedAt      int64  `json:"created_at"`
}
