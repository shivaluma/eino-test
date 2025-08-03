package handlers

import (
	"net/http"
	"strconv"

	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/models"
	"github.com/shivaluma/eino-agent/internal/repository"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ConversationHandler struct {
	convRepo *repository.ConversationRepository
	authSvc  *auth.Service
}

func NewConversationHandler(convRepo *repository.ConversationRepository, authSvc *auth.Service) *ConversationHandler {
	return &ConversationHandler{
		convRepo: convRepo,
		authSvc:  authSvc,
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

func (h *ConversationHandler) CreateConversation(c echo.Context) error {
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req models.CreateConversationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	conversation := &models.Conversation{
		UserID: userClaims.UserID,
		Title:  req.Title,
	}

	if err := h.convRepo.Create(c.Request().Context(), conversation); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create conversation",
		})
	}

	return c.JSON(http.StatusCreated, conversation)
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