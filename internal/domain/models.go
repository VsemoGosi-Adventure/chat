package domain

import (
	"time"

	"github.com/google/uuid"
)

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	ID             uuid.UUID  `json:"id"`
	ChatID         uuid.UUID  `json:"chat_id"`
	UserID         uuid.UUID  `json:"user_id"`
	SenderUsername string     `json:"sender_username"`
	ReplyToID      *uuid.UUID `json:"reply_to_id,omitempty"`
	Type           MessageType `json:"type"`
	Body           string     `json:"body"`
	IsDeleted      bool       `json:"is_deleted"`
	CreatedAt      time.Time  `json:"created_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
}

type Chat struct {
	ID            uuid.UUID  `json:"id"`
	EventID       uuid.UUID  `json:"event_id"`
	IsActive      bool       `json:"is_active"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type IncomingMessage struct {
	ReplyToID *uuid.UUID  `json:"reply_to_id,omitempty"`
	Type      MessageType `json:"type"`
	Body      string      `json:"body"`
}

type OutgoingMessage struct {
	Message
	EventType string `json:"event_type"`
}

const (
	EventTypeNewMessage    = "message.new"
	EventTypeDeleteMessage = "message.deleted"
)
