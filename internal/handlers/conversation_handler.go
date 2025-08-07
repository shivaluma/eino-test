package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/shivaluma/eino-agent/internal/ai"
	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/models"
	"github.com/shivaluma/eino-agent/internal/repository"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ConversationHandler struct {
	convRepo  *repository.ConversationRepository
	authSvc   *auth.Service
	aiService ai.Service
}

func NewConversationHandler(convRepo *repository.ConversationRepository, authSvc *auth.Service, aiService ai.Service) *ConversationHandler {
	return &ConversationHandler{
		convRepo:  convRepo,
		authSvc:   authSvc,
		aiService: aiService,
	}
}

func (h *ConversationHandler) GetConversations(c echo.Context) error {
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	limit := 20
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	conversations, err := h.convRepo.GetByUserID(c.Request().Context(), userClaims.UserID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch conversations",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"conversations": conversations,
		"limit":         limit,
		"offset":        offset,
	})
}

func (h *ConversationHandler) SendMessage(c echo.Context) error {
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req models.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()
	var conversation *models.Conversation
	var chatHistory []*schema.Message

	// Check if conversation exists or create new one
	if req.ConversationID != nil {
		// Existing conversation
		conversation, err = h.convRepo.GetByID(ctx, *req.ConversationID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch conversation",
			})
		}
		if conversation == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Conversation not found",
			})
		}
		if conversation.UserID != userClaims.UserID {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Access denied",
			})
		}

		// Load chat history
		messages, err := h.convRepo.GetMessages(ctx, conversation.ID, 50, 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch messages",
			})
		}

		// Convert to schema messages for chat history
		for _, msg := range messages {
			switch msg.SenderType {
			case models.SenderTypeUser:
				chatHistory = append(chatHistory, schema.UserMessage(msg.Content))
			case models.SenderTypeAgent:
				chatHistory = append(chatHistory, schema.AssistantMessage(msg.Content, nil))
			}
		}
	} else {
		// New conversation - generate title from first message
		title, err := h.aiService.GenerateTitle(ctx, req.Message)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate title",
			})
		}

		conversation = &models.Conversation{
			UserID: userClaims.UserID,
			Title:  &title,
		}

		if err := h.convRepo.Create(ctx, conversation); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create conversation",
			})
		}
	}

	// Save user message
	userMessage := &models.Message{
		ConversationID: conversation.ID,
		SenderID:       userClaims.UserID,
		SenderType:     models.SenderTypeUser,
		Content:        req.Message,
		Metadata:       req.Metadata,
	}

	if err := h.convRepo.CreateMessage(ctx, userMessage); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save message",
		})
	}

	// Update conversation's updated_at
	if err := h.convRepo.UpdateTimestamp(ctx, conversation.ID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update conversation timestamp: %v\n", err)
	}

	// Prepare AI request
	aiRequest := &ai.ChatRequest{
		Message:        req.Message,
		ConversationID: conversation.ID.String(),
		UserID:         userClaims.UserID.String(),
		Stream:         req.Stream,
		History:        chatHistory,
	}

	// Handle streaming or regular response
	if req.Stream {
		// Set headers for chunked streaming
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		// Write initial response with conversation and message info
		initialData := map[string]interface{}{
			"conversation_id": conversation.ID,
			"message_id":      userMessage.ID,
			"type":            "init",
		}
		initialJSON, _ := json.Marshal(initialData)
		c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", string(initialJSON))))
		c.Response().Flush()

		// Stream callback
		streamCallback := func(chunk string) error {
			chunkData := map[string]interface{}{
				"type":    "chunk",
				"content": chunk,
			}
			chunkJSON, _ := json.Marshal(chunkData)
			_, err := c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", string(chunkJSON))))
			if err != nil {
				return err // Client disconnected
			}
			c.Response().Flush()
			return nil
		}

		// Stream the response
		response, err := h.aiService.Stream(ctx, aiRequest, streamCallback)
		if err != nil {
			errorData := map[string]interface{}{
				"type":  "error",
				"error": err.Error(),
			}
			errorJSON, _ := json.Marshal(errorData)
			c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", string(errorJSON))))
			c.Response().Flush()
			return nil
		}

		fullContent := response.Content

		// Save AI response
		aiMessage := &models.Message{
			ConversationID: conversation.ID,
			SenderID:       uuid.Nil, // System/AI doesn't have a user ID
			SenderType:     models.SenderTypeAgent,
			Content:        fullContent,
		}

		if err := h.convRepo.CreateMessage(ctx, aiMessage); err != nil {
			// Log error but don't fail the streaming
			fmt.Printf("Failed to save AI message: %v\n", err)
		}

		// Send completion signal
		completeData := map[string]interface{}{
			"type":       "complete",
			"message_id": aiMessage.ID,
		}
		completeJSON, _ := json.Marshal(completeData)
		c.Response().Write([]byte(fmt.Sprintf("data: %s\n\n", string(completeJSON))))
		c.Response().Flush()

		return nil
	} else {
		// Non-streaming response
		response, err := h.aiService.Generate(ctx, aiRequest)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate response",
			})
		}

		// Save AI response
		aiMessage := &models.Message{
			ConversationID: conversation.ID,
			SenderID:       uuid.Nil,
			SenderType:     models.SenderTypeAgent,
			Content:        response.Content,
		}

		if err := h.convRepo.CreateMessage(ctx, aiMessage); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to save AI response",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"conversation_id": conversation.ID,
			"user_message":    userMessage,
			"ai_message":      aiMessage,
		})
	}
}

func (h *ConversationHandler) StreamMessage(c echo.Context) error {
	return h.SendMessage(c)
}

func (h *ConversationHandler) GetConversation(c echo.Context) error {
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid conversation ID",
		})
	}

	conversation, err := h.convRepo.GetByID(c.Request().Context(), conversationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch conversation",
		})
	}
	if conversation == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Conversation not found",
		})
	}

	if conversation.UserID != userClaims.UserID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(http.StatusOK, conversation)
}

func (h *ConversationHandler) GetMessages(c echo.Context) error {
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid conversation ID",
		})
	}

	conversation, err := h.convRepo.GetByID(c.Request().Context(), conversationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch conversation",
		})
	}
	if conversation == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Conversation not found",
		})
	}

	if conversation.UserID != userClaims.UserID {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	limit := 50
	offset := 0

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := h.convRepo.GetMessages(c.Request().Context(), conversationID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch messages",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// Deprecated - use SendMessage instead
func (h *ConversationHandler) CreateConversation(c echo.Context) error {
	return h.SendMessage(c)
}
