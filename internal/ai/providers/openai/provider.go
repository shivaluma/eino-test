package openai

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/shivaluma/eino-agent/internal/ai"
)

// Provider implements the AI Provider interface for OpenAI
type Provider struct {
	config *Config
}

// Config holds OpenAI-specific configuration
type Config struct {
	APIKey    string
	BaseURL   string
	Model     string
	OrgID     string
	Timeout   int
	MaxTokens int
}

// NewProvider creates a new OpenAI provider
func NewProvider() ai.Provider {
	return &Provider{
		config: loadConfigFromEnv(),
	}
}

// NewProviderWithConfig creates a new OpenAI provider with custom config
func NewProviderWithConfig(config *Config) ai.Provider {
	return &Provider{
		config: config,
	}
}

func loadConfigFromEnv() *Config {
	return &Config{
		APIKey:    os.Getenv("OPENAI_API_KEY"),
		BaseURL:   os.Getenv("OPENAI_BASE_URL"),
		Model:     getEnvOrDefault("OPENAI_MODEL_NAME", "gpt-4.1-mini"),
		OrgID:     os.Getenv("OPENAI_ORG_ID"),
		MaxTokens: 2000,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateChatModel creates an OpenAI chat model instance
func (p *Provider) CreateChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	if !p.IsAvailable() {
		return nil, fmt.Errorf("OpenAI provider is not available: missing API key")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: p.config.BaseURL,
		Model:   p.config.Model,
		APIKey:  p.config.APIKey,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}

	return chatModel, nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "openai"
}

// IsAvailable checks if the provider is properly configured
func (p *Provider) IsAvailable() bool {
	return p.config.APIKey != ""
}

// GetModel returns the configured model name
func (p *Provider) GetModel() string {
	return p.config.Model
}

// UpdateConfig updates the provider configuration
func (p *Provider) UpdateConfig(config *Config) {
	p.config = config
}
