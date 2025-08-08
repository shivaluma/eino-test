package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/models"
	"github.com/shivaluma/eino-agent/internal/repository"
	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	userRepo    *repository.UserRepository
	oauthRepo   *repository.OAuthRepository
	authSvc     *auth.Service
	oauthSvc    *auth.OAuthService
	frontendURL string
}

func NewOAuthHandler(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	authSvc *auth.Service,
	oauthSvc *auth.OAuthService,
	frontendURL string,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:    userRepo,
		oauthRepo:   oauthRepo,
		authSvc:     authSvc,
		oauthSvc:    oauthSvc,
		frontendURL: frontendURL,
	}
}

// InitiateOAuth initiates the OAuth flow for the specified provider
func (h *OAuthHandler) InitiateOAuth(c echo.Context) error {
	provider := c.Param("provider")

	if !h.oauthSvc.IsProviderEnabled(provider) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Provider %s is not enabled", provider),
		})
	}

	// Generate state for CSRF protection
	state, err := h.oauthSvc.GenerateState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate state",
		})
	}

	// Store state in database with expiration
	oauthState := &models.OAuthState{
		State:     state,
		Provider:  provider,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	// For PKCE (optional, mainly for mobile apps)
	if c.QueryParam("pkce") == "true" {
		verifier, challenge, err := h.oauthSvc.GeneratePKCE()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate PKCE",
			})
		}
		oauthState.CodeVerifier = &verifier

		// Get auth URL with PKCE challenge
		authURL, err := h.oauthSvc.GetAuthURL(
			provider,
			state,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate authorization URL",
			})
		}

		if err := h.oauthRepo.StoreState(c.Request().Context(), oauthState); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to store OAuth state",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"auth_url": authURL,
			"state":    state,
		})
	}

	// Regular OAuth flow without PKCE
	authURL, err := h.oauthSvc.GetAuthURL(provider, state)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authorization URL",
		})
	}

	if err := h.oauthRepo.StoreState(c.Request().Context(), oauthState); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to store OAuth state",
		})
	}

	// For web flow, redirect directly
	if c.QueryParam("redirect") != "false" {
		return c.Redirect(http.StatusTemporaryRedirect, authURL)
	}

	// For API/mobile flow, return the URL
	return c.JSON(http.StatusOK, map[string]string{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleOAuthCallback handles the OAuth callback from the provider
func (h *OAuthHandler) HandleOAuthCallback(c echo.Context) error {
	provider := c.Param("provider")

	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")

	// Handle OAuth errors
	if errorParam != "" {
		errorDesc := c.QueryParam("error_description")
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=%s&error_description=%s",
			h.frontendURL, errorParam, errorDesc)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Validate state
	if state == "" {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=invalid_state", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Retrieve and validate state from database
	storedState, err := h.oauthRepo.GetState(c.Request().Context(), state)
	if err != nil || storedState == nil {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=invalid_state", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Check state expiration
	if time.Now().After(storedState.ExpiresAt) {
		h.oauthRepo.DeleteState(c.Request().Context(), state)
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=state_expired", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Delete state after validation (one-time use)
	defer h.oauthRepo.DeleteState(c.Request().Context(), state)

	// Exchange code for tokens
	var opts []oauth2.AuthCodeOption
	if storedState.CodeVerifier != nil {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", *storedState.CodeVerifier))
	}

	token, err := h.oauthSvc.ExchangeCode(c.Request().Context(), provider, code, opts...)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=token_exchange_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Get user info from provider
	userInfo, err := h.oauthSvc.GetUserInfo(c.Request().Context(), provider, token)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=user_info_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Check if OAuth account exists
	oauthAccount, err := h.oauthRepo.GetByProviderID(c.Request().Context(), provider, userInfo.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check OAuth account",
		})
	}

	var user *models.User

	if oauthAccount != nil {
		// Existing OAuth account - get the user
		user, err = h.userRepo.GetByID(c.Request().Context(), oauthAccount.UserID)
		if err != nil || user == nil {
			redirectURL := fmt.Sprintf("%s/auth/sign-in?error=user_not_found", h.frontendURL)
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		}

		// Update OAuth account tokens
		oauthAccount.AccessToken = &token.AccessToken
		if token.RefreshToken != "" {
			oauthAccount.RefreshToken = &token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			oauthAccount.TokenExpiresAt = &token.Expiry
		}

		// Update user data
		userDataJSON, _ := json.Marshal(userInfo)
		oauthAccount.RawUserData = userDataJSON

		if err := h.oauthRepo.UpdateAccount(c.Request().Context(), oauthAccount); err != nil {
			// Non-critical error, log but continue
			fmt.Printf("Failed to update OAuth account: %v\n", err)
		}
	} else {
		// New OAuth account - check if user with email exists
		if userInfo.Email != "" {
			existingUser, _ := h.userRepo.GetByEmail(c.Request().Context(), userInfo.Email)
			if existingUser != nil {
				// Link OAuth account to existing user
				user = existingUser
			}
		}

		if user == nil {
			// Create new user
			username := userInfo.Username
			if username == "" {
				username = strings.Split(userInfo.Email, "@")[0]
			}

			// Ensure unique username
			baseUsername := username
			for i := 1; ; i++ {
				existingUser, _ := h.userRepo.GetByUsername(c.Request().Context(), username)
				if existingUser == nil {
					break
				}
				username = fmt.Sprintf("%s%d", baseUsername, i)
			}

			user = &models.User{
				Username:        username,
				Email:           userInfo.Email,
				OAuthProvider:   &provider,
				OAuthProviderID: &userInfo.ID,
				AvatarURL:       &userInfo.AvatarURL,
				OAuthEmail:      &userInfo.Email,
			}

			if err := h.userRepo.Create(c.Request().Context(), user); err != nil {
				redirectURL := fmt.Sprintf("%s/auth/sign-in?error=user_creation_failed", h.frontendURL)
				return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			}
		}

		// Create OAuth account
		userDataJSON, _ := json.Marshal(userInfo)
		oauthAccount = &models.OAuthAccount{
			UserID:            user.ID,
			Provider:          provider,
			ProviderAccountID: userInfo.ID,
			ProviderEmail:     &userInfo.Email,
			ProviderUsername:  &userInfo.Username,
			ProviderAvatarURL: &userInfo.AvatarURL,
			AccessToken:       &token.AccessToken,
			RawUserData:       userDataJSON,
		}

		if token.RefreshToken != "" {
			oauthAccount.RefreshToken = &token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			oauthAccount.TokenExpiresAt = &token.Expiry
		}

		if err := h.oauthRepo.CreateAccount(c.Request().Context(), oauthAccount); err != nil {
			// Non-critical error if user was created successfully
			fmt.Printf("Failed to create OAuth account: %v\n", err)
		}
	}

	// Generate JWT tokens
	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=token_generation_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		redirectURL := fmt.Sprintf("%s/auth/sign-in?error=token_generation_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Store refresh token
	refreshTokenRecord := h.authSvc.CreateRefreshTokenRecord(user.ID, refreshToken)
	if err := h.userRepo.StoreRefreshToken(c.Request().Context(), refreshTokenRecord); err != nil {
		// Non-critical error
		fmt.Printf("Failed to store refresh token: %v\n", err)
	}

	// Redirect to frontend with tokens
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		h.frontendURL, accessToken, refreshToken)

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GetOAuthProviders returns the list of enabled OAuth providers
func (h *OAuthHandler) GetOAuthProviders(c echo.Context) error {
	providers := h.oauthSvc.GetEnabledProviders()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"providers": providers,
	})
}

// LinkOAuthAccount links an OAuth account to an existing user
func (h *OAuthHandler) LinkOAuthAccount(c echo.Context) error {
	// Get user from context (requires authentication)
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	provider := c.Param("provider")

	if !h.oauthSvc.IsProviderEnabled(provider) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Provider %s is not enabled", provider),
		})
	}

	// Generate state with user ID embedded
	state, err := h.oauthSvc.GenerateState()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate state",
		})
	}

	// Store state with user context
	redirectURI := fmt.Sprintf("/account/linked?provider=%s", provider)
	oauthState := &models.OAuthState{
		State:       state,
		Provider:    provider,
		RedirectURI: &redirectURI,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	if err := h.oauthRepo.StoreState(c.Request().Context(), oauthState); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to store OAuth state",
		})
	}

	// Store user ID in session/cookie for linking after callback
	c.SetCookie(&http.Cookie{
		Name:     "oauth_link_user",
		Value:    userClaims.UserID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	authURL, err := h.oauthSvc.GetAuthURL(provider, state)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authorization URL",
		})
	}

	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// UnlinkOAuthAccount unlinks an OAuth account from a user
func (h *OAuthHandler) UnlinkOAuthAccount(c echo.Context) error {
	// Get user from context (requires authentication)
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	provider := c.Param("provider")

	// Check if user has other auth methods
	user, err := h.userRepo.GetByID(c.Request().Context(), userClaims.UserID)
	if err != nil || user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	// Count OAuth accounts
	accounts, err := h.oauthRepo.GetByUserID(c.Request().Context(), userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get OAuth accounts",
		})
	}

	// Ensure user has another auth method
	if len(accounts) <= 1 && user.PasswordHash == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Cannot unlink the only authentication method",
		})
	}

	// Delete OAuth account
	if err := h.oauthRepo.DeleteByUserAndProvider(c.Request().Context(), userClaims.UserID, provider); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to unlink OAuth account",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "OAuth account unlinked successfully",
	})
}

// GetLinkedAccounts returns the list of linked OAuth accounts for a user
func (h *OAuthHandler) GetLinkedAccounts(c echo.Context) error {
	// Get user from context (requires authentication)
	userClaims, err := h.authSvc.GetUserClaimsFromContext(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	accounts, err := h.oauthRepo.GetByUserID(c.Request().Context(), userClaims.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get linked accounts",
		})
	}

	// Filter sensitive data
	var linkedAccounts []map[string]interface{}
	for _, account := range accounts {
		linkedAccount := map[string]interface{}{
			"provider":   account.Provider,
			"username":   account.ProviderUsername,
			"email":      account.ProviderEmail,
			"avatar_url": account.ProviderAvatarURL,
			"created_at": account.CreatedAt,
		}
		linkedAccounts = append(linkedAccounts, linkedAccount)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"linked_accounts": linkedAccounts,
	})
}
