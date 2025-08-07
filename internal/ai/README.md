# AI Package Architecture

## Overview
This package provides a clean, extensible architecture for AI/LLM integration in the application.

## Package Structure

```
ai/
├── service.go          # Main AI service interface and implementation
├── types.go            # Common types and interfaces
├── providers/          # AI provider implementations (OpenAI, Claude, etc.)
│   ├── factory.go      # Provider factory pattern
│   └── openai/         # OpenAI-specific implementation
└── templates/          # Message template management
    └── manager.go      # Template manager with configuration
```

## Design Principles

### 1. **Separation of Concerns**
- **Service Layer**: Business logic and orchestration
- **Provider Layer**: Vendor-specific implementations
- **Template Layer**: Message formatting and prompt engineering

### 2. **Dependency Inversion**
- Core service depends on interfaces, not concrete implementations
- Easy to swap providers without changing business logic

### 3. **Factory Pattern**
- Centralized provider creation and management
- Runtime provider selection based on configuration

### 4. **Template Management**
- Externalized prompt templates
- Configurable system prompts and behaviors
- Future: Load templates from files/database

## Usage Example

```go
// Initialize provider factory
factory := providers.NewFactory()

// Get a provider (OpenAI in this case)
provider, err := factory.GetProvider(providers.OpenAI)
if err != nil {
    log.Fatal(err)
}

// Create chat model
model, err := provider.CreateChatModel(ctx)
if err != nil {
    log.Fatal(err)
}

// Create AI service
aiService := ai.NewService(model, &ai.Config{
    DefaultModel: "gpt-3.5-turbo",
    SystemPrompt: "You are a helpful assistant",
})

// Use the service
response, err := aiService.Generate(ctx, &ai.ChatRequest{
    Message: "Hello, how are you?",
    UserID:  "user123",
})
```

## Adding New Providers

1. Create a new package under `providers/` (e.g., `providers/anthropic/`)
2. Implement the `ai.Provider` interface
3. Register in the factory (`factory.go`)

Example:
```go
// providers/anthropic/provider.go
type Provider struct {
    config *Config
}

func (p *Provider) CreateChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
    // Implementation
}

func (p *Provider) GetName() string {
    return "anthropic"
}

func (p *Provider) IsAvailable() bool {
    return p.config.APIKey != ""
}
```

## Benefits

1. **Maintainability**: Clear separation of concerns
2. **Testability**: Easy to mock interfaces for testing
3. **Extensibility**: Add new providers without changing existing code
4. **Configuration**: Centralized configuration management
5. **Error Handling**: Consistent error handling across providers

## Future Enhancements

1. **Template Loading**: Load templates from YAML/JSON files
2. **Provider Health Checks**: Periodic availability checks
3. **Rate Limiting**: Per-provider rate limiting
4. **Metrics**: Provider usage metrics and monitoring
5. **Caching**: Response caching for similar queries
6. **Fallback**: Automatic fallback to alternative providers