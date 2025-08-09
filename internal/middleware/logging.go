package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shivaluma/eino-agent/internal/logger"
)

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get or generate request ID
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = logger.GenerateRequestID()
			}

			// Add to context
			ctx := logger.WithRequestID(c.Request().Context(), requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Add to response header
			c.Response().Header().Set("X-Request-ID", requestID)

			return next(c)
		}
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Get request ID
			requestID := logger.GetRequestID(c.Request().Context())

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Calculate latency
			latency := time.Since(start)

			// Get response status
			status := c.Response().Status

			// Log the request
			log := logger.WithContext(c.Request().Context())
			
			fields := map[string]interface{}{
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"status":     status,
				"latency_ms": latency.Milliseconds(),
				"ip":         c.RealIP(),
				"user_agent": c.Request().UserAgent(),
			}

			if requestID != "" {
				fields["request_id"] = requestID
			}

			// Add query parameters if present
			if c.Request().URL.RawQuery != "" {
				fields["query"] = c.Request().URL.RawQuery
			}

			// Add error if present
			if err != nil {
				fields["error"] = err.Error()
			}

			// Log based on status code
			event := log.With().Fields(fields).Logger()
			
			switch {
			case status >= 500:
				event.Error().Msg("Server error")
			case status >= 400:
				event.Warn().Msg("Client error")
			case status >= 300:
				event.Info().Msg("Redirection")
			default:
				event.Info().Msg("Request completed")
			}

			return nil
		}
	}
}

// ErrorHandlingMiddleware handles errors and logs them
func ErrorHandlingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			// Get logger with context
			log := logger.WithContext(c.Request().Context())

			// Handle Echo HTTP errors
			if he, ok := err.(*echo.HTTPError); ok {
				log.Warn().
					Int("status", he.Code).
					Interface("message", he.Message).
					Msg("HTTP error")
				return err
			}

			// Log other errors
			log.Error().
				Err(err).
				Str("path", c.Request().URL.Path).
				Str("method", c.Request().Method).
				Msg("Unhandled error")

			return err
		}
	}
}