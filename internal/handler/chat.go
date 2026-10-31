package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/VsemoGosi-Adventure/chat/internal/service"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

type GetMessagesResponse struct {
	Messages []MessageDTO `json:"messages"`
	HasMore  bool         `json:"has_more"`
}

type MessageDTO struct {
	ID             uuid.UUID  `json:"id"`
	ChatID         uuid.UUID  `json:"chat_id"`
	UserID         uuid.UUID  `json:"user_id"`
	SenderUsername string     `json:"sender_username"`
	ReplyToID      *uuid.UUID `json:"reply_to_id,omitempty"`
	Type           string     `json:"type"`
	Body           string     `json:"body"`
	CreatedAt      time.Time  `json:"created_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
}

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("event_id")
	if eventIDStr == "" {
		http.Error(w, "Missing event_id", http.StatusBadRequest)
		return
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		http.Error(w, "Invalid event_id format", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		if parsedCursor, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			cursor = &parsedCursor
		}
	}

	chat, err := h.chatService.GetOrCreateChat(r.Context(), eventID)
	if err != nil {
		http.Error(w, "Failed to get chat", http.StatusInternalServerError)
		return
	}

	messages, err := h.chatService.GetMessages(r.Context(), chat.ID, cursor, limit+1)
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	messageDTOs := make([]MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTOs = append(messageDTOs, MessageDTO{
			ID:             msg.ID,
			ChatID:         msg.ChatID,
			UserID:         msg.UserID,
			SenderUsername: msg.SenderUsername,
			ReplyToID:      msg.ReplyToID,
			Type:           string(msg.Type),
			Body:           msg.Body,
			CreatedAt:      msg.CreatedAt,
			EditedAt:       msg.EditedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetMessagesResponse{
		Messages: messageDTOs,
		HasMore:  hasMore,
	})
}
