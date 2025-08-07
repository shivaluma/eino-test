package handlers

import (
	"net/http"
	"strings"

	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/models"
	"github.com/shivaluma/eino-agent/internal/repository"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	authSvc  *auth.Service
}

func NewAuthHandler(userRepo *repository.UserRepository, authSvc *auth.Service) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		authSvc:  authSvc,
	}
}

func (h *AuthHandler) CheckEmail(c echo.Context) error {
	var req models.CheckEmailRequest
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

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existingUser, err := h.userRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"exists": existingUser != nil,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.UserRegisterRequest
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

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	existingUser, err := h.userRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}
	if existingUser != nil {
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "Email already exists",
		})
	}

	hashedPassword, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process password",
		})
	}

	user := &models.User{
		Username:     req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := h.userRepo.Create(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.UserLoginRequest
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

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.userRepo.GetByEmail(c.Request().Context(), req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	if err := h.authSvc.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid email or password",
		})
	}

	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	refreshTokenRecord := h.authSvc.CreateRefreshTokenRecord(user.ID, refreshToken)
	if err := h.userRepo.StoreRefreshToken(c.Request().Context(), refreshTokenRecord); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to store refresh token",
		})
	}

	return c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req models.RefreshTokenRequest
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

	refreshTokenRecord, err := h.userRepo.GetRefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}
	if refreshTokenRecord == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired refresh token",
		})
	}

	user, err := h.userRepo.GetByID(c.Request().Context(), refreshTokenRecord.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not found",
		})
	}

	if err := h.userRepo.InvalidateRefreshToken(c.Request().Context(), refreshTokenRecord.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to invalidate refresh token",
		})
	}

	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	newRefreshToken, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	newRefreshTokenRecord := h.authSvc.CreateRefreshTokenRecord(user.ID, newRefreshToken)
	if err := h.userRepo.StoreRefreshToken(c.Request().Context(), newRefreshTokenRecord); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to store refresh token",
		})
	}

	return c.JSON(http.StatusOK, models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}