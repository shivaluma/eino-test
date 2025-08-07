package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     *string   `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID             int64           `json:"id" db:"id"`
	ConversationID uuid.UUID       `json:"conversation_id" db:"conversation_id"`
	SenderID       uuid.UUID       `json:"sender_id" db:"sender_id"`
	SenderType     string          `json:"sender_type" db:"sender_type"`
	Content        string          `json:"content" db:"content"`
	Metadata       json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

type SendMessageRequest struct {
	Message        string          `json:"message" validate:"required"`
	ConversationID *uuid.UUID      `json:"conversation_id,omitempty"`
	Model          string          `json:"model,omitempty"`
	Stream         bool            `json:"stream"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
}

type CreateMessageRequest struct {
	Content  string          `json:"content" validate:"required"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type ConversationWithMessages struct {
	Conversation
	Messages []Message `json:"messages"`
}

type MessageResponse struct {
	ID             int64           `json:"id"`
	ConversationID uuid.UUID       `json:"conversation_id"`
	SenderID       uuid.UUID       `json:"sender_id"`
	SenderType     string          `json:"sender_type"`
	Content        string          `json:"content"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

const (
	SenderTypeUser  = "USER"
	SenderTypeAgent = "AGENT"
)
