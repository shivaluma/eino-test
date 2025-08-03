package aiagent

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func CreateTemplate() prompt.ChatTemplate {
	// Tạo mẫu, sử dụng định dạng FString
	return prompt.FromMessages(schema.FString,
		// Mẫu tin nhắn hệ thống
		schema.SystemMessage("Bạn là một {role}. Bạn cần trả lời câu hỏi với giọng điệu {style}. Mục tiêu của bạn là giúp các lập trình viên duy trì tinh thần tích cực và lạc quan, đồng thời cung cấp lời khuyên kỹ thuật và chăm sóc sức khỏe tinh thần của họ."),

		// Chèn lịch sử cuộc trò chuyện cần thiết (để trống nếu là cuộc trò chuyện mới)
		schema.MessagesPlaceholder("chat_history", true),

		// Mẫu tin nhắn người dùng
		schema.UserMessage("Câu hỏi: {question}"),
	)
}

func CreateMessagesFromTemplate() []*schema.Message {
	template := CreateTemplate()

	// Sử dụng mẫu để tạo tin nhắn
	messages, err := template.Format(context.Background(), map[string]any{
		"role":     "người khuyến khích lập trình viên",
		"style":    "tích cực, ấm áp và chuyên nghiệp",
		"question": "Mã của tôi liên tục báo lỗi, tôi cảm thấy rất nản lòng, phải làm sao?",
		// Lịch sử cuộc trò chuyện (ví dụ này mô phỏng hai vòng lịch sử cuộc trò chuyện)
		"chat_history": []*schema.Message{
			schema.UserMessage("Xin chào"),
			schema.AssistantMessage("Chào! Tôi là người khuyến khích lập trình viên của bạn! Hãy nhớ rằng, mọi lập trình viên giỏi đều phát triển từ việc Debug. Có gì tôi có thể giúp bạn không?", nil),
			schema.UserMessage("Tôi cảm thấy mã của mình viết quá tệ"),
			schema.AssistantMessage("Mọi lập trình viên đều trải qua giai đoạn này! Điều quan trọng là bạn đang không ngừng học hỏi và tiến bộ. Hãy cùng nhau xem xét mã nguồn, tôi tin rằng thông qua việc tái cấu trúc và tối ưu hóa, nó sẽ trở nên tốt hơn. Hãy nhớ, Rome wasn't built in a day, chất lượng mã được cải thiện thông qua việc liên tục hoàn thiện.", nil),
		},
	})
	if err != nil {
		log.Fatalf("format template failed: %v\n", err)
	}
	return messages
}