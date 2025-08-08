package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Username         string     `json:"username" db:"username"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     *string    `json:"-" db:"password_hash"` // Nullable for OAuth-only users
	OAuthProvider    *string    `json:"oauth_provider,omitempty" db:"oauth_provider"`
	OAuthProviderID  *string    `json:"-" db:"oauth_provider_id"`
	AvatarURL        *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	OAuthEmail       *string    `json:"-" db:"oauth_email"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type UserRegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type CheckEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty" db:"used_at"`
}

type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// OAuth-specific models
type OAuthAccount struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	Provider           string     `json:"provider" db:"provider"`
	ProviderAccountID  string     `json:"provider_account_id" db:"provider_account_id"`
	ProviderEmail      *string    `json:"provider_email,omitempty" db:"provider_email"`
	ProviderUsername   *string    `json:"provider_username,omitempty" db:"provider_username"`
	ProviderAvatarURL  *string    `json:"provider_avatar_url,omitempty" db:"provider_avatar_url"`
	AccessToken        *string    `json:"-" db:"access_token"`
	RefreshToken       *string    `json:"-" db:"refresh_token"`
	TokenExpiresAt     *time.Time `json:"-" db:"token_expires_at"`
	RawUserData        []byte     `json:"-" db:"raw_user_data"` // JSONB
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type OAuthState struct {
	ID           uuid.UUID `json:"id" db:"id"`
	State        string    `json:"state" db:"state"`
	Provider     string    `json:"provider" db:"provider"`
	CodeVerifier *string   `json:"-" db:"code_verifier"` // For PKCE
	RedirectURI  *string   `json:"redirect_uri,omitempty" db:"redirect_uri"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
}