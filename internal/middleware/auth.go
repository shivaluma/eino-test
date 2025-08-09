package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/shivaluma/eino-agent/internal/auth"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authSvc *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header required",
				})
			}

			tokenParts := strings.SplitN(authHeader, " ", 2)
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format",
				})
			}

			token, err := authSvc.ValidateAccessToken(tokenParts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired token",
				})
			}

			userID, err := authSvc.ExtractUserIDFromToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			username, err := authSvc.ExtractUsernameFromToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			ctx := context.WithValue(c.Request().Context(), "user_id", userID)
			ctx = context.WithValue(ctx, "username", username)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func CORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// In production, consider restricting origin. For now, reflect request origin for cookies.
			origin := c.Request().Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			c.Response().Header().Set("Vary", "Origin")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}
