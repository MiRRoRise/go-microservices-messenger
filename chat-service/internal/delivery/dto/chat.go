package dto

type CreateChatRequest struct {
    Name string `json:"name" binding:"required"`
}

type CreateChatResponse struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}

type ChatResponse struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}

type ListChatsResponse struct {
    Chats []ChatResponse `json:"chats"`
}