package dto

type CreateMessageRequest struct {
	Text string `json:"text"`
}

type CreateMessageResponse struct {
	ID        int64  `json:"id"`
	ChatID    int64  `json:"chat_id"`
	SenderID  int64  `json:"sender_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type MessageResponse struct {
	ID        int64  `json:"id"`
	ChatID    int64  `json:"chat_id"`
	SenderID  int64  `json:"sender_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type ListMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}
