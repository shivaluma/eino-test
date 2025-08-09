package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/logger"
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

	log := logger.WithContext(c.Request().Context())
	log.Debug().Str("provider", provider).Msg("Exchanging code for tokens")
	token, err := h.oauthSvc.ExchangeCode(c.Request().Context(), provider, code, opts...)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("Token exchange failed")
		redirectURL := fmt.Sprintf("%s/sign-in?error=token_exchange_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}
	log.Debug().
		Str("provider", provider).
		Bool("has_access_token", token.AccessToken != "").
		Bool("has_refresh_token", token.RefreshToken != "").
		Msg("Token exchange successful")

	// Get user info from provider
	log.Debug().Str("provider", provider).Msg("Getting user info from provider")
	userInfo, err := h.oauthSvc.GetUserInfo(c.Request().Context(), provider, token)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("Failed to get user info")
		redirectURL := fmt.Sprintf("%s/sign-in?error=user_info_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Validate essential OAuth user info
	if userInfo == nil {
		log.Error().Str("provider", provider).Msg("OAuth provider returned nil user info")
		redirectURL := fmt.Sprintf("%s/sign-in?error=invalid_user_data", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	if userInfo.ID == "" {
		log.Error().Str("provider", provider).Msg("OAuth provider returned empty user ID")
		redirectURL := fmt.Sprintf("%s/sign-in?error=invalid_user_id", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Sanitize and validate user data
	userInfo.ID = strings.TrimSpace(userInfo.ID)
	userInfo.Email = strings.TrimSpace(strings.ToLower(userInfo.Email))
	userInfo.Username = strings.TrimSpace(userInfo.Username)

	// Validate email format if provided
	if userInfo.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(userInfo.Email) {
			log.Warn().
				Str("provider", provider).
				Str("invalid_email", userInfo.Email).
				Msg("OAuth provider returned invalid email format - clearing email")
			userInfo.Email = ""
		}
	}

	log.Debug().
		Str("provider", provider).
		Str("user_id", userInfo.ID).
		Str("email", userInfo.Email).
		Str("username", userInfo.Username).
		Bool("has_avatar", userInfo.AvatarURL != "").
		Msg("User info retrieved and validated")

	// Check if OAuth account exists
	log.Debug().
		Str("provider", provider).
		Str("provider_id", userInfo.ID).
		Msg("Checking if OAuth account exists")
	oauthAccount, err := h.oauthRepo.GetByProviderID(c.Request().Context(), provider, userInfo.ID)
	if err != nil {
		// Only treat actual database errors as errors, not "no rows found"
		log.Error().
			Err(err).
			Str("provider", provider).
			Str("provider_id", userInfo.ID).
			Msg("Database error while checking OAuth account")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error during authentication",
		})
	}

	if oauthAccount != nil {
		log.Debug().
			Interface("user_id", oauthAccount.UserID).
			Msg("Existing OAuth account found - proceeding with login flow")
	} else {
		log.Debug().
			Str("provider", provider).
			Str("provider_id", userInfo.ID).
			Msg("No OAuth account found - proceeding with new user registration flow")
	}

	var user *models.User

	if oauthAccount != nil {
		// Existing OAuth account - get the user
		log.Debug().Interface("user_id", oauthAccount.UserID).Msg("Existing OAuth account found")
		user, err = h.userRepo.GetByID(c.Request().Context(), oauthAccount.UserID)
		if err != nil || user == nil {
			log.Error().
				Err(err).
				Interface("user_id", oauthAccount.UserID).
				Msg("Failed to get user")
			redirectURL := fmt.Sprintf("%s/sign-in?error=user_not_found", h.frontendURL)
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		}
		log.Debug().
			Str("username", user.Username).
			Str("email", user.Email).
			Msg("User found")

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
			log.Warn().Err(err).Msg("Failed to update OAuth account")
		}
	} else {
		// New OAuth account - check if user with email exists
		log.Info().
			Str("provider", provider).
			Str("provider_id", userInfo.ID).
			Str("email", userInfo.Email).
			Msg("Starting new user registration flow")

		if userInfo.Email != "" {
			log.Debug().Str("email", userInfo.Email).Msg("Checking if user exists with this email address")
			existingUser, err := h.userRepo.GetByEmail(c.Request().Context(), userInfo.Email)
			if err != nil {
				log.Warn().
					Err(err).
					Str("email", userInfo.Email).
					Msg("Error checking for existing user by email - proceeding with new user creation")
			}
			if existingUser != nil {
				// Link OAuth account to existing user
				log.Info().
					Interface("user_id", existingUser.ID).
					Str("username", existingUser.Username).
					Str("email", userInfo.Email).
					Msg("Found existing user with same email - linking OAuth account")
				user = existingUser
			}
		} else {
			log.Info().Msg("OAuth provider did not return email address - creating user without email linking")
		}

		if user == nil {
			// Create new user
			log.Info().
				Str("provider", provider).
				Str("oauth_username", userInfo.Username).
				Str("oauth_email", userInfo.Email).
				Msg("Creating completely new user account")

			username := userInfo.Username
			if username == "" {
				if userInfo.Email != "" {
					username = strings.Split(userInfo.Email, "@")[0]
					log.Debug().
						Str("generated_username", username).
						Msg("Generated username from email address")
				} else {
					// Ensure we have at least 8 characters for the ID substring
					idSubstr := userInfo.ID
					if len(userInfo.ID) > 8 {
						idSubstr = userInfo.ID[:8]
					}
					username = fmt.Sprintf("%s_user_%s", provider, idSubstr)
					log.Debug().
						Str("generated_username", username).
						Msg("Generated fallback username from provider and ID")
				}
			}

			// Sanitize username: remove invalid characters, ensure it's not too long
			username = strings.TrimSpace(username)
			username = regexp.MustCompile(`[^a-zA-Z0-9_\-.]`).ReplaceAllString(username, "_")
			if len(username) > 50 {
				username = username[:50]
			}
			if username == "" || username == "_" {
				// Ultimate fallback if username becomes empty after sanitization
				username = fmt.Sprintf("%s_user_%d", provider, time.Now().Unix()%10000)
				log.Warn().
					Str("fallback_username", username).
					Msg("Used timestamp-based fallback username due to sanitization")
			}

			log.Debug().
				Str("final_username_candidate", username).
				Msg("Username sanitized and ready for uniqueness check")

			// Ensure unique username with proper error handling and limits
			baseUsername := username
			maxAttempts := 100 // Prevent infinite loops
			for i := 1; i <= maxAttempts; i++ {
				log.Debug().
					Str("checking_username", username).
					Int("attempt", i).
					Msg("Checking username availability")

				existingUser, err := h.userRepo.GetByUsername(c.Request().Context(), username)
				if err != nil {
					log.Warn().
						Err(err).
						Str("username", username).
						Msg("Error checking username availability - continuing anyway")
					// Continue with the username if DB check fails
					break
				}
				if existingUser == nil {
					log.Debug().
						Str("final_username", username).
						Int("attempts_taken", i).
						Msg("Found available username")
					break
				}

				if i == maxAttempts {
					// Fallback to UUID-based username if we can't find a unique one
					username = fmt.Sprintf("%s_%s_user", provider, userInfo.ID[:8])
					log.Warn().
						Str("fallback_username", username).
						Str("base_username", baseUsername).
						Msg("Reached max attempts for username generation, using fallback")
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

			log.Debug().
				Str("username", user.Username).
				Str("email", user.Email).
				Str("provider", provider).
				Str("provider_id", userInfo.ID).
				Msg("Starting atomic user and OAuth account creation")

			log.Debug().
				Str("username", user.Username).
				Str("email", user.Email).
				Str("provider", provider).
				Str("provider_id", userInfo.ID).
				Msg("Creating user")
			if err := h.userRepo.Create(c.Request().Context(), user); err != nil {
				log.Error().Err(err).Msg("Failed to create user")
				redirectURL := fmt.Sprintf("%s/sign-in?error=user_creation_failed", h.frontendURL)
				return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			}
			log.Debug().Interface("user_id", user.ID).Msg("User created successfully")
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

		log.Debug().
			Interface("user_id", user.ID).
			Str("provider", provider).
			Str("provider_id", userInfo.ID).
			Msg("Creating OAuth account")

		if err := h.oauthRepo.CreateAccount(c.Request().Context(), oauthAccount); err != nil {
			log.Error().Err(err).Msg("Failed to create OAuth account")
			redirectURL := fmt.Sprintf("%s/sign-in?error=oauth_account_creation_failed", h.frontendURL)
			return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		}
		log.Debug().Msg("OAuth account created successfully")
	}

	// Generate JWT tokens
	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/sign-in?error=token_generation_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken()
	if err != nil {
		redirectURL := fmt.Sprintf("%s/sign-in?error=token_generation_failed", h.frontendURL)
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// Store refresh token
	refreshTokenRecord := h.authSvc.CreateRefreshTokenRecord(user.ID, refreshToken)
	if err := h.userRepo.StoreRefreshToken(c.Request().Context(), refreshTokenRecord); err != nil {
		// Non-critical error
		fmt.Printf("Failed to store refresh token: %v\n", err)
	}

	// Set secure HTTP-only cookies for tokens (following world best practices)
	// Access token cookie - shorter expiration
	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   15 * 60, // 15 minutes
	})

	// Refresh token cookie - longer expiration
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Redirect to frontend OAuth callback for client-side handling
	redirectURL := fmt.Sprintf("%s/oauth/callback?success=true", h.frontendURL)
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
