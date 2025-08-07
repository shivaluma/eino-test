package ai

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ChatRequest represents a request to the AI chat service
type ChatRequest struct {
	Message        string
	ConversationID string
	UserID         string
	Model          string
	Stream         bool
	History        []*schema.Message
}

// ChatResponse represents a response from the AI chat service
type ChatResponse struct {
	Content        string
	ConversationID string
	MessageID      int64
}

// StreamCallback is called for each chunk in streaming mode
type StreamCallback func(chunk string) error

// Service defines the interface for AI chat operations
type Service interface {
	// Generate creates a single response
	Generate(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	
	// Stream creates a streaming response
	Stream(ctx context.Context, req *ChatRequest, callback StreamCallback) (*ChatResponse, error)
	
	// GenerateTitle generates a title for a conversation
	GenerateTitle(ctx context.Context, firstMessage string) (string, error)
}

// Provider defines the interface for AI model providers
type Provider interface {
	CreateChatModel(ctx context.Context) (model.ToolCallingChatModel, error)
	GetName() string
	IsAvailable() bool
}

// Config holds AI service configuration
type Config struct {
	DefaultModel    string
	DefaultProvider string
	SystemPrompt    string
	Temperature     float64
	MaxTokens       int
}