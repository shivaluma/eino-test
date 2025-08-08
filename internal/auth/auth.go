package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/shivaluma/eino-agent/config"
	"github.com/shivaluma/eino-agent/internal/models"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	config *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{config: cfg}
}

func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) VerifyPassword(hashedPassword *string, password string) error {
	if hashedPassword == nil {
		return fmt.Errorf("password authentication not available for this account")
	}
	return bcrypt.CompareHashAndPassword([]byte(*hashedPassword), []byte(password))
}

func (s *Service) GenerateAccessToken(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	token, err := jwt.NewBuilder().
		Issuer("food-agent").
		Subject(userID.String()).
		Audience([]string{"food-agent-api"}).
		IssuedAt(now).
		Expiration(now.Add(s.config.JWT.AccessExpiration)).
		Claim("username", username).
		Claim("type", "access").
		Build()

	if err != nil {
		return "", fmt.Errorf("failed to build access token: %w", err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, []byte(s.config.JWT.AccessSecret)))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return string(signed), nil
}

func (s *Service) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *Service) CreateRefreshTokenRecord(userID uuid.UUID, token string) *models.RefreshToken {
	return &models.RefreshToken{
		UserID:    userID,
		TokenHash: token,
		ExpiresAt: time.Now().Add(s.config.JWT.RefreshExpiration),
	}
}

func (s *Service) ValidateAccessToken(tokenString string) (jwt.Token, error) {
	token, err := jwt.Parse([]byte(tokenString), jwt.WithKey(jwa.HS256, []byte(s.config.JWT.AccessSecret)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	if err := jwt.Validate(token); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	tokenType, ok := token.Get("type")
	if !ok || tokenType != "access" {
		return nil, fmt.Errorf("invalid token type")
	}

	return token, nil
}

func (s *Service) ExtractUserIDFromToken(token jwt.Token) (uuid.UUID, error) {
	subject := token.Subject()
	if subject == "" {
		return uuid.Nil, fmt.Errorf("no subject in token")
	}

	userID, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

func (s *Service) ExtractUsernameFromToken(token jwt.Token) (string, error) {
	username, ok := token.Get("username")
	if !ok {
		return "", fmt.Errorf("no username in token")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("invalid username format")
	}

	return usernameStr, nil
}

type UserClaims struct {
	UserID   uuid.UUID
	Username string
}

func (s *Service) GetUserClaimsFromContext(ctx context.Context) (*UserClaims, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return nil, fmt.Errorf("user ID not found in context")
	}

	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, fmt.Errorf("username not found in context")
	}

	return &UserClaims{
		UserID:   userID,
		Username: username,
	}, nil
}