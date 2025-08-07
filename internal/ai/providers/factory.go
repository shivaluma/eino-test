package providers

import (
	"fmt"

	"github.com/shivaluma/eino-agent/internal/ai"
	"github.com/shivaluma/eino-agent/internal/ai/providers/openai"
)

// ProviderType represents the type of AI provider
type ProviderType string

const (
	OpenAI    ProviderType = "openai"
	Anthropic ProviderType = "anthropic"
	Gemini    ProviderType = "gemini"
)

// Factory creates AI providers based on type
type Factory struct {
	providers map[ProviderType]ai.Provider
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	f := &Factory{
		providers: make(map[ProviderType]ai.Provider),
	}

	// Register default providers
	f.Register(OpenAI, openai.NewProvider())

	// Future: Register other providers
	// f.Register(Anthropic, anthropic.NewProvider())
	// f.Register(Gemini, gemini.NewProvider())

	return f
}

// Register registers a new provider
func (f *Factory) Register(providerType ProviderType, provider ai.Provider) {
	f.providers[providerType] = provider
}

// GetProvider returns a provider by type
func (f *Factory) GetProvider(providerType ProviderType) (ai.Provider, error) {
	provider, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}

	if !provider.IsAvailable() {
		return nil, fmt.Errorf("provider %s is not available", providerType)
	}

	return provider, nil
}

// GetAvailableProviders returns all available providers
func (f *Factory) GetAvailableProviders() []string {
	var available []string
	for providerType, provider := range f.providers {
		if provider.IsAvailable() {
			available = append(available, string(providerType))
		}
	}
	return available
}

// GetDefaultProvider returns the first available provider
func (f *Factory) GetDefaultProvider() (ai.Provider, error) {
	// Priority order
	priority := []ProviderType{OpenAI, Anthropic, Gemini}

	for _, providerType := range priority {
		if provider, err := f.GetProvider(providerType); err == nil {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("no available providers found")
}
