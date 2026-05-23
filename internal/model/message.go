package model

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	ReceiverID     string `json:"receiver_id"`
	Content        string `json:"content"`
	Type           string `json:"type"`
}

// SendMessageResponse 发送消息响应体
type SendMessageResponse struct {
	MessageID string `json:"message_id"`
	CreatedAt string `json:"created_at"`
}
