package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MiRRoRise/chat-service/internal/delivery/dto"
	"github.com/MiRRoRise/chat-service/internal/domain"
	"github.com/MiRRoRise/chat-service/internal/usecase"
	"github.com/MiRRoRise/chat-service/pkg/jwt"
	"github.com/MiRRoRise/chat-service/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	chatUseCase    usecase.ChatUseCase
	messageUseCase usecase.MessageUseCase
	tokenManager   jwt.TokenManager
}

func NewHandler(chatUseCase usecase.ChatUseCase, messageUseCase usecase.MessageUseCase, tokenManager jwt.TokenManager) *Handler {
	return &Handler{
		chatUseCase:    chatUseCase,
		messageUseCase: messageUseCase,
		tokenManager:   tokenManager,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// CreateChat godoc
// @Summary Create a new chat
// @Description Create a new chat with the given name
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatRequest true "Chat name"
// @Success 201 {object} dto.CreateChatResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /chats [post]
func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "error decode json", http.StatusBadRequest)
		return
	}

	chat, err := h.chatUseCase.CreateChat(r.Context(), req.Name)
	if err != nil {
		switch err {
		case domain.ErrChatNameRequired:
			Error(w, "chat name required", http.StatusBadRequest)
		case domain.ErrChatExists:
			Error(w, "chat already exists", http.StatusConflict)
		default:
			Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.CreateChatResponse{
		ID:        chat.ID,
		Name:      chat.Name,
		CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	JSON(w, resp, http.StatusCreated)
}

// ListChats godoc
// @Summary Get all chats
// @Description Get a list of all chats
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ListChatsResponse
// @Failure 401 {object} map[string]string
// @Router /chats [get]
func (h *Handler) ListChats(w http.ResponseWriter, r *http.Request) {
	chats, err := h.chatUseCase.ListChats(r.Context())
	if err != nil {
		Error(w, "failed to list chats", http.StatusInternalServerError)
		return
	}

	resp := dto.ListChatsResponse{
		Chats: make([]dto.ChatResponse, len(chats)),
	}

	for i, chat := range chats {
		resp.Chats[i] = dto.ChatResponse{
			ID:        chat.ID,
			Name:      chat.Name,
			CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	JSON(w, resp, http.StatusOK)
}

// GetChatByID godoc
// @Summary Get chat by ID
// @Description Get a chat by its ID
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Success 200 {object} dto.ChatResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /chats/{id} [get]
func (h *Handler) GetChatByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	chatID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	chat, err := h.chatUseCase.GetChatByID(r.Context(), chatID)
	if err != nil {
		switch err {
		case domain.ErrChatIDRequired:
			Error(w, "chat id required", http.StatusBadRequest)
		case domain.ErrChatNotFound:
			Error(w, "chat not found", http.StatusNotFound)
		default:
			Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.ChatResponse{
		ID:        chat.ID,
		Name:      chat.Name,
		CreatedAt: chat.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	JSON(w, resp, http.StatusOK)
}

// CreateMessage godoc
// @Summary Send a message
// @Description Send a message to a chat
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Param request body dto.CreateMessageRequest true "Message text"
// @Success 201 {object} dto.CreateMessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /chats/{id}/messages [post]
func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	chatID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	var req dto.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, "error decode json", http.StatusBadRequest)
		return
	}

	senderID, err := GetIDFromContext(r.Context())
	if err != nil {
		Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	message, err := h.messageUseCase.CreateMessage(r.Context(), chatID, senderID, req.Text)
	if err != nil {
		switch err {
		case domain.ErrMessageEmpty:
			Error(w, "message cannot be empty", http.StatusBadRequest)
		case domain.ErrChatNotFound:
			Error(w, "chat not found", http.StatusNotFound)
		default:
			Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.CreateMessageResponse{
		ID:        message.ID,
		ChatID:    message.ChatID,
		SenderID:  message.SenderID,
		Text:      message.Text,
		CreatedAt: message.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	JSON(w, resp, http.StatusCreated)
}

// ListMessages godoc
// @Summary Get chat messages
// @Description Get all messages from a chat
// @Tags messages
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Success 200 {object} dto.ListMessagesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /chats/{id}/messages [get]
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	chatID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	messages, err := h.messageUseCase.ListMessages(r.Context(), chatID)
	if err != nil {
		switch err {
		case domain.ErrChatNotFound:
			Error(w, "chat not found", http.StatusNotFound)
		default:
			Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	resp := dto.ListMessagesResponse{
		Messages: make([]dto.MessageResponse, len(messages)),
	}

	for i, message := range messages {
		resp.Messages[i] = dto.MessageResponse{
			ID:        message.ID,
			ChatID:    message.ChatID,
			SenderID:  message.SenderID,
			Text:      message.Text,
			CreatedAt: message.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	JSON(w, resp, http.StatusOK)
}

func JSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.New().Error("error encode json: %w",  err)
		}
	}
}

func Error(w http.ResponseWriter, message string, status int) {
	JSON(w, map[string]string{"error": message}, status)
}
