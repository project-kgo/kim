package event

type MessageEvent struct {
	MessageID      int64  `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id"`
	Content        string `json:"content"`
	Type           string `json:"type"`
	CreatedAt      int64  `json:"created_at"`
}
