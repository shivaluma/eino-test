package repository

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/shivaluma/eino-agent/internal/database"
	"github.com/shivaluma/eino-agent/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, oauth_provider, oauth_provider_id, avatar_url, oauth_email)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.OAuthProvider,
		user.OAuthProviderID,
		user.AvatarURL,
		user.OAuthEmail,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, oauth_provider, oauth_provider_id, avatar_url, oauth_email, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.OAuthProvider, &user.OAuthProviderID, &user.AvatarURL, &user.OAuthEmail,
			&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, oauth_provider, oauth_provider_id, avatar_url, oauth_email, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.OAuthProvider, &user.OAuthProviderID, &user.AvatarURL, &user.OAuthEmail,
			&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, oauth_provider, oauth_provider_id, avatar_url, oauth_email, created_at, updated_at
		FROM users
		WHERE username = $1`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.OAuthProvider, &user.OAuthProviderID, &user.AvatarURL, &user.OAuthEmail,
			&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) StoreRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	tokenHash := sha256.Sum256([]byte(token.TokenHash))
	token.TokenHash = fmt.Sprintf("%x", tokenHash)

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	return r.db.Pool.QueryRow(ctx, query, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, tokenString string) (*models.RefreshToken, error) {
	tokenHash := sha256.Sum256([]byte(tokenString))
	hashedToken := fmt.Sprintf("%x", tokenHash)

	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, used_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()`

	token := &models.RefreshToken{}
	err := r.db.Pool.QueryRow(ctx, query, hashedToken).
		Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.UsedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

func (r *UserRepository) InvalidateRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET used_at = NOW()
		WHERE id = $1`

	_, err := r.db.Pool.Exec(ctx, query, tokenID)
	return err
}

func (r *UserRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR used_at IS NOT NULL`

	_, err := r.db.Pool.Exec(ctx, query)
	return err
}
