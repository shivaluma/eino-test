package templates

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// Manager manages AI message templates
type Manager struct {
	chatTemplate          prompt.ChatTemplate
	titleTemplate         prompt.ChatTemplate
	foodRecommendTemplate prompt.ChatTemplate
	config                *Config
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

// FoodRecommendConfig returns configuration for food recommendation agent
func FoodRecommendConfig() *Config {
	return &Config{
		Role:       "Food Expert & Culinary Advisor",
		Style:      "thân thiện, hài hước và chuyên nghiệp về ẩm thực",
		Language:   "Vietnamese",
		MaxHistory: 5, // More history for better food context
	}
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return NewManagerWithConfig(DefaultConfig())
}

// NewFoodRecommendManager creates a new template manager for food recommendations
func NewFoodRecommendManager() *Manager {
	return NewManagerWithConfig(FoodRecommendConfig())
}

// NewManagerWithConfig creates a new template manager with custom config
func NewManagerWithConfig(config *Config) *Manager {
	return &Manager{
		chatTemplate:          createChatTemplate(),
		titleTemplate:         createTitleTemplate(),
		foodRecommendTemplate: createFoodRecommendTemplate(),
		config:                config,
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

func createFoodRecommendTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage(`Bạn là một AI agent chuyên nghiệp, thân thiện và có chút hài hước về ẩm thực. Giao tiếp tự nhiên, gần gũi như một người bạn am hiểu ẩm thực.

Mục tiêu: Đề xuất món ăn hấp dẫn, cung cấp các tùy chọn đa dạng, và gợi mở để tiếp tục cuộc trò chuyện.

Ngôn ngữ: Sử dụng ngôn từ trẻ trung, tích cực, ví dụ: "đỉnh của chóp," "thần thánh," "quốc dân." Kết hợp biểu tượng cảm xúc (emoji) để tăng tính tương tác.

Cấu trúc phản hồi:

1. Phản ứng ban đầu (Warm-up): Bày tỏ sự hào hứng với lựa chọn của người dùng, sử dụng các câu cảm thán.

2. Gợi ý các biến thể (Options): Đưa ra từ 2-3 tùy chọn hấp dẫn liên quan đến món ăn mà người dùng đã chọn. Mỗi tùy chọn cần có mô tả ngắn gọn, sinh động để kích thích vị giác.

3. Câu hỏi mở (Open-ended question): Kết thúc bằng một câu hỏi để người dùng có thể lựa chọn hoặc yêu cầu thêm thông tin, giúp duy trì cuộc hội thoại.

Ví dụ với yêu cầu "bún bò":
- Bắt đầu: "Ố là la! Một sự lựa chọn không thể tuyệt vời hơn! 🍜✨"
- Gợi ý: "Bún bò truyền thống với nước lèo đậm đà, chả cua thơm phức 🦀, hoặc bún bò giò heo với giò heo hầm mềm béo ngậy 🥩"
- Kết thúc: "Bạn ưng ý 'em' nào trong danh sách trên, hay muốn tôi gợi ý thêm vài quán bún bò 'thần thánh' gần bạn? 😋."`),
		schema.MessagesPlaceholder("chat_history", true),
		schema.UserMessage("{food_request}"),
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

// BuildFoodRecommendMessages builds messages for food recommendation
func (m *Manager) BuildFoodRecommendMessages(foodRequest string, history []*schema.Message) ([]*schema.Message, error) {
	// Limit history to configured max
	if len(history) > m.config.MaxHistory*2 { // *2 because each exchange has user + assistant
		history = history[len(history)-m.config.MaxHistory*2:]
	}

	params := map[string]any{
		"food_request": foodRequest,
	}

	// Only add chat_history if it exists
	if len(history) > 0 {
		params["chat_history"] = history
	}

	messages, err := m.foodRecommendTemplate.Format(context.Background(), params)

	if err != nil {
		return nil, fmt.Errorf("failed to format food recommendation template: %w", err)
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
