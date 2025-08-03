package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/shivaluma/eino-agent/internal/aiagent"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	ctx := context.Background()

	// 使用模版创建messages
	log.Printf("===create messages===\n")
	messages := aiagent.CreateMessagesFromTemplate()
	log.Printf("messages: %+v\n\n", messages)

	// 创建llm
	log.Printf("===create llm===\n")
	cm := aiagent.CreateOpenAIChatModel(ctx)
	// cm := createOllamaChatModel(ctx)
	log.Printf("create llm success\n\n")

	log.Printf("===llm generate===\n")
	result := aiagent.Generate(ctx, cm, messages)
	log.Printf("result: %+v\n\n", result)

	log.Printf("===llm stream generate===\n")
	streamResult := aiagent.Stream(ctx, cm, messages)
	aiagent.ReportStream(streamResult)
}
