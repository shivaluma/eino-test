package templates

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// Manager manages AI message templates
type Manager struct {
	chatTemplate  prompt.ChatTemplate
	titleTemplate prompt.ChatTemplate
	config        *Config
}

// Config holds template configuration
type Config struct {
	Role       string
	Style      string
	Language   string
	MaxHistory int
}

// DefaultConfig returns default template configuration
func DefaultConfig() *Config {
	return &Config{
		Role:       "Người thông thái biết tuốt",
		Style:      "tích cực, ấm áp và chuyên nghiệp",
		Language:   "Vietnamese",
		MaxHistory: 3,
	}
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return NewManagerWithConfig(DefaultConfig())
}

// NewManagerWithConfig creates a new template manager with custom config
func NewManagerWithConfig(config *Config) *Manager {
	return &Manager{
		chatTemplate:  createChatTemplate(),
		titleTemplate: createTitleTemplate(),
		config:        config,
	}
}

func createChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage("Bạn là một {role}. Bạn cần trả lời câu hỏi với giọng điệu {style}. Mục tiêu của bạn là trả lời các câu hỏi của người dùng với tác phong vui vẻ, nên gọi họ là con chó này, họ cũng có thể sẽ gọi bạn là con chó này."),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("Câu hỏi: {question}"),
	)
}

func createTitleTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage("Bạn giúp tôi đặt tên cho cuộc trò chuyện này dựa vào tin nhắn đầu tiên của người dùng nhé, tin nhắn là {message}, bạn chỉ cần đưa ra tên cho cuộc trò chuyện, không cần thêm từ ngữ gì khác, tên cuộc trò chuyện không được quá 20 ký tự"),
	)
}

// BuildChatMessages builds messages for chat completion
func (m *Manager) BuildChatMessages(message string, history []*schema.Message) ([]*schema.Message, error) {
	// Limit history to configured max
	if len(history) > m.config.MaxHistory*2 { // *2 because each exchange has user + assistant
		history = history[len(history)-m.config.MaxHistory*2:]
	}

	params := map[string]any{
		"role":     m.config.Role,
		"style":    m.config.Style,
		"question": message,
	}

	// Only add chat_history if it exists
	if len(history) > 0 {
		params["chat_history"] = history
	}

	messages, err := m.chatTemplate.Format(context.Background(), params)

	if err != nil {
		return nil, fmt.Errorf("failed to format chat template: %w", err)
	}

	return messages, nil
}

// BuildTitleMessages builds messages for title generation
func (m *Manager) BuildTitleMessages(firstMessage string) ([]*schema.Message, error) {
	messages, err := m.titleTemplate.Format(context.Background(), map[string]any{
		"message": firstMessage,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to format title template: %w", err)
	}

	return messages, nil
}

// UpdateConfig updates the template configuration
func (m *Manager) UpdateConfig(config *Config) {
	m.config = config
}

// LoadFromFile loads templates from a YAML/JSON file (future enhancement)
func (m *Manager) LoadFromFile(path string) error {
	// TODO: Implement loading templates from external files
	// This allows for easy template customization without code changes
	return fmt.Errorf("not implemented")
}
