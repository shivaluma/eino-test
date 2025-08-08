package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shivaluma/eino-agent/config"
	"github.com/shivaluma/eino-agent/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	config    *config.Config
	providers map[string]*oauth2.Config
}

func NewOAuthService(cfg *config.Config) *OAuthService {
	providers := make(map[string]*oauth2.Config)

	if cfg.OAuth.GitHub.Enabled {
		providers["github"] = &oauth2.Config{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			RedirectURL:  cfg.OAuth.GitHub.RedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	if cfg.OAuth.Google.Enabled {
		providers["google"] = &oauth2.Config{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}
	}

	return &OAuthService{
		config:    cfg,
		providers: providers,
	}
}

// GenerateState generates a secure random state parameter for OAuth flow
func (s *OAuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GeneratePKCE generates code verifier and challenge for PKCE flow
func (s *OAuthService) GeneratePKCE() (verifier string, challenge string, err error) {
	// Generate code verifier
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate code verifier: %w", err)
	}
	verifier = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)

	// Generate code challenge
	h := sha256.New()
	h.Write([]byte(verifier))
	challenge = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h.Sum(nil))

	return verifier, challenge, nil
}

// GetAuthURL returns the OAuth authorization URL for the specified provider
func (s *OAuthService) GetAuthURL(provider, state string, opts ...oauth2.AuthCodeOption) (string, error) {
	cfg, exists := s.providers[provider]
	if !exists {
		return "", fmt.Errorf("provider %s not configured or enabled", provider)
	}

	// Add state to options
	opts = append(opts, oauth2.SetAuthURLParam("state", state))

	// Add provider-specific parameters
	if provider == "google" {
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "select_account"))
		opts = append(opts, oauth2.AccessTypeOffline)
	}

	return cfg.AuthCodeURL(state, opts...), nil
}

// ExchangeCode exchanges the authorization code for tokens
func (s *OAuthService) ExchangeCode(ctx context.Context, provider, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	cfg, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured or enabled", provider)
	}

	token, err := cfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// GetUserInfo fetches user information from the OAuth provider
func (s *OAuthService) GetUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	switch provider {
	case "github":
		return s.getGitHubUserInfo(ctx, token)
	case "google":
		return s.getGoogleUserInfo(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *OAuthService) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := s.providers["github"].Client(ctx, token)

	// Get user info
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", body)
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If email is not public, fetch from emails endpoint
	if githubUser.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()

			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}

			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err == nil {
				for _, e := range emails {
					if e.Primary && e.Verified {
						githubUser.Email = e.Email
						break
					}
				}
			}
		}
	}

	// Use login as name if name is empty
	name := githubUser.Name
	if name == "" {
		name = githubUser.Login
	}

	return &models.OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     githubUser.Email,
		Name:      name,
		Username:  githubUser.Login,
		AvatarURL: githubUser.AvatarURL,
		Provider:  "github",
	}, nil
}

func (s *OAuthService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*models.OAuthUserInfo, error) {
	client := s.providers["google"].Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", body)
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// Extract username from email
	username := strings.Split(googleUser.Email, "@")[0]

	return &models.OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		Username:  username,
		AvatarURL: googleUser.Picture,
		Provider:  "google",
	}, nil
}

// ValidateState validates the OAuth state parameter
func (s *OAuthService) ValidateState(state string) error {
	if state == "" {
		return fmt.Errorf("state parameter is required")
	}

	// Additional validation can be added here
	// For example, checking against stored states in database

	return nil
}

// IsProviderEnabled checks if a provider is enabled
func (s *OAuthService) IsProviderEnabled(provider string) bool {
	_, exists := s.providers[provider]
	return exists
}

// GetEnabledProviders returns a list of enabled OAuth providers
func (s *OAuthService) GetEnabledProviders() []string {
	providers := make([]string, 0, len(s.providers))
	for provider := range s.providers {
		providers = append(providers, provider)
	}
	return providers
}
