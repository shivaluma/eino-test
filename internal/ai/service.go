package ai

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/shivaluma/eino-agent/internal/ai/templates"
)

type service struct {
	model     model.ToolCallingChatModel
	templates *templates.Manager
	config    *Config
}

// NewService creates a new AI service
func NewService(model model.ToolCallingChatModel, config *Config) Service {
	return &service{
		model:     model,
		templates: templates.NewManager(),
		config:    config,
	}
}

func (s *service) Generate(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Build messages with template
	messages, err := s.templates.BuildFoodRecommendMessages(req.Message, req.History)
	if err != nil {
		return nil, fmt.Errorf("failed to build messages: %w", err)
	}

	// Generate response
	response, err := s.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return &ChatResponse{
		Content:        response.Content,
		ConversationID: req.ConversationID,
	}, nil
}

func (s *service) Stream(ctx context.Context, req *ChatRequest, callback StreamCallback) (*ChatResponse, error) {
	// Build messages with template
	messages, err := s.templates.BuildFoodRecommendMessages(req.Message, req.History)
	if err != nil {
		return nil, fmt.Errorf("failed to build messages: %w", err)
	}

	// Start streaming
	streamReader, err := s.model.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}

	var fullContent string
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			if err == schema.ErrRecvAfterClosed {
				break
			}
			return nil, fmt.Errorf("stream error: %w", err)
		}

		if chunk != nil && chunk.Content != "" {
			fullContent += chunk.Content
			if err := callback(chunk.Content); err != nil {
				return nil, fmt.Errorf("callback error: %w", err)
			}
		}
	}

	return &ChatResponse{
		Content:        fullContent,
		ConversationID: req.ConversationID,
	}, nil
}

func (s *service) GenerateTitle(ctx context.Context, firstMessage string) (string, error) {
	messages, err := s.templates.BuildTitleMessages(firstMessage)
	if err != nil {
		return "", fmt.Errorf("failed to build title messages: %w", err)
	}

	response, err := s.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate title: %w", err)
	}

	return response.Content, nil
}
