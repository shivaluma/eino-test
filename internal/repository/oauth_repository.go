package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shivaluma/eino-agent/internal/models"
)

type OAuthRepository struct {
	db *pgxpool.Pool
}

func NewOAuthRepository(db *pgxpool.Pool) *OAuthRepository {
	return &OAuthRepository{db: db}
}

// StoreState stores an OAuth state for CSRF protection
func (r *OAuthRepository) StoreState(ctx context.Context, state *models.OAuthState) error {
	query := `
		INSERT INTO oauth_states (state, provider, code_verifier, redirect_uri, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, 
		state.State, 
		state.Provider, 
		state.CodeVerifier, 
		state.RedirectURI, 
		state.ExpiresAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store OAuth state: %w", err)
	}

	return nil
}

// GetState retrieves an OAuth state by its value
func (r *OAuthRepository) GetState(ctx context.Context, state string) (*models.OAuthState, error) {
	query := `
		SELECT id, state, provider, code_verifier, redirect_uri, expires_at, created_at
		FROM oauth_states
		WHERE state = $1 AND expires_at > NOW()
		LIMIT 1
	`

	var oauthState models.OAuthState
	err := r.db.QueryRow(ctx, query, state).Scan(
		&oauthState.ID,
		&oauthState.State,
		&oauthState.Provider,
		&oauthState.CodeVerifier,
		&oauthState.RedirectURI,
		&oauthState.ExpiresAt,
		&oauthState.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth state: %w", err)
	}

	return &oauthState, nil
}

// DeleteState deletes an OAuth state
func (r *OAuthRepository) DeleteState(ctx context.Context, state string) error {
	query := `DELETE FROM oauth_states WHERE state = $1`
	
	_, err := r.db.Exec(ctx, query, state)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth state: %w", err)
	}

	return nil
}

// CreateAccount creates a new OAuth account
func (r *OAuthRepository) CreateAccount(ctx context.Context, account *models.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (
			user_id, provider, provider_account_id, provider_email, 
			provider_username, provider_avatar_url, access_token, 
			refresh_token, token_expires_at, raw_user_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		account.UserID,
		account.Provider,
		account.ProviderAccountID,
		account.ProviderEmail,
		account.ProviderUsername,
		account.ProviderAvatarURL,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.RawUserData,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create OAuth account: %w", err)
	}

	return nil
}

// GetByProviderID gets an OAuth account by provider and provider account ID
func (r *OAuthRepository) GetByProviderID(ctx context.Context, provider, providerAccountID string) (*models.OAuthAccount, error) {
	query := `
		SELECT 
			id, user_id, provider, provider_account_id, provider_email,
			provider_username, provider_avatar_url, access_token,
			refresh_token, token_expires_at, raw_user_data, created_at, updated_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_account_id = $2
		LIMIT 1
	`

	var account models.OAuthAccount
	err := r.db.QueryRow(ctx, query, provider, providerAccountID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderAccountID,
		&account.ProviderEmail,
		&account.ProviderUsername,
		&account.ProviderAvatarURL,
		&account.AccessToken,
		&account.RefreshToken,
		&account.TokenExpiresAt,
		&account.RawUserData,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth account: %w", err)
	}

	return &account, nil
}

// GetByUserID gets all OAuth accounts for a user
func (r *OAuthRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.OAuthAccount, error) {
	query := `
		SELECT 
			id, user_id, provider, provider_account_id, provider_email,
			provider_username, provider_avatar_url, access_token,
			refresh_token, token_expires_at, raw_user_data, created_at, updated_at
		FROM oauth_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*models.OAuthAccount
	for rows.Next() {
		var account models.OAuthAccount
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderAccountID,
			&account.ProviderEmail,
			&account.ProviderUsername,
			&account.ProviderAvatarURL,
			&account.AccessToken,
			&account.RefreshToken,
			&account.TokenExpiresAt,
			&account.RawUserData,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}

// UpdateAccount updates an OAuth account
func (r *OAuthRepository) UpdateAccount(ctx context.Context, account *models.OAuthAccount) error {
	query := `
		UPDATE oauth_accounts
		SET 
			provider_email = $2,
			provider_username = $3,
			provider_avatar_url = $4,
			access_token = $5,
			refresh_token = $6,
			token_expires_at = $7,
			raw_user_data = $8,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		account.ID,
		account.ProviderEmail,
		account.ProviderUsername,
		account.ProviderAvatarURL,
		account.AccessToken,
		account.RefreshToken,
		account.TokenExpiresAt,
		account.RawUserData,
	)

	if err != nil {
		return fmt.Errorf("failed to update OAuth account: %w", err)
	}

	return nil
}

// DeleteByUserAndProvider deletes an OAuth account for a user and provider
func (r *OAuthRepository) DeleteByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) error {
	query := `DELETE FROM oauth_accounts WHERE user_id = $1 AND provider = $2`
	
	_, err := r.db.Exec(ctx, query, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth account: %w", err)
	}

	return nil
}

// CleanupExpiredStates removes expired OAuth states
func (r *OAuthRepository) CleanupExpiredStates(ctx context.Context) error {
	query := `DELETE FROM oauth_states WHERE expires_at < NOW()`
	
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired states: %w", err)
	}

	return nil
}