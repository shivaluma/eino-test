package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivaluma/eino-agent/config"
	"github.com/shivaluma/eino-agent/internal/ai"
	"github.com/shivaluma/eino-agent/internal/ai/providers"
	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/database"
	"github.com/shivaluma/eino-agent/internal/handlers"
	"github.com/shivaluma/eino-agent/internal/logger"
	"github.com/shivaluma/eino-agent/internal/middleware"
	"github.com/shivaluma/eino-agent/internal/migrations"
	"github.com/shivaluma/eino-agent/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	// Initialize logger based on environment
	logConfig := &logger.Config{
		Level:           getEnvOrDefault("LOG_LEVEL", "info"),
		Format:          getEnvOrDefault("LOG_FORMAT", "json"),
		Output:          getEnvOrDefault("LOG_OUTPUT", "stdout"),
		FilePath:        getEnvOrDefault("LOG_FILE_PATH", "logs/app.log"),
		AddTimestamp:    true,
		AddCaller:       true,
		PrettyPrint:     cfg.Environment == "development",
		ErrorStackTrace: true,
	}

	if cfg.Environment == "development" {
		logConfig.Level = "debug"
		logConfig.Format = "console"
		logConfig.PrettyPrint = true
	}

	if err := logger.Init(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// From now on, use structured logging
	logger.Info("Starting Eino Agent server")
	logger.WithField("environment", cfg.Environment).Info("Configuration loaded")

	db, err := database.New(cfg)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Run database migrations on startup
	logger.Info("Running database migrations...")
	migrator := migrations.NewMigrator(db.Pool, "migrations", cfg)
	if err := migrator.Migrate(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}
	logger.Info("Database migrations completed successfully")

	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	oauthRepo := repository.NewOAuthRepository(db.Pool)
	authSvc := auth.NewService(cfg)
	oauthSvc := auth.NewOAuthService(cfg)

	// Initialize AI service with provider factory
	ctx := context.Background()
	factory := providers.NewFactory()
	provider, err := factory.GetDefaultProvider()
	if err != nil {
		logger.WithError(err).Fatal("Failed to get AI provider")
	}

	model, err := provider.CreateChatModel(ctx)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create chat model")
	}

	aiService := ai.NewService(model, &ai.Config{
		DefaultProvider: provider.GetName(),
	})

	authHandler := handlers.NewAuthHandler(userRepo, authSvc)
	oauthHandler := handlers.NewOAuthHandler(userRepo, oauthRepo, authSvc, oauthSvc, cfg.OAuth.FrontendURL)
	convHandler := handlers.NewConversationHandler(convRepo, authSvc, aiService)

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	// Add request ID middleware first
	e.Use(middleware.RequestIDMiddleware())
	// Replace Echo's logger with our structured logger
	e.Use(middleware.LoggingMiddleware())
	e.Use(middleware.ErrorHandlingMiddleware())
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORSMiddleware())

	api := e.Group("/api/v1")

	api.POST("/check-email", authHandler.CheckEmail)
	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	api.POST("/token/refresh", authHandler.RefreshToken)

	// OAuth routes
	api.GET("/auth/oauth/providers", oauthHandler.GetOAuthProviders)
	api.GET("/auth/oauth/:provider/authorize", oauthHandler.InitiateOAuth)
	api.GET("/auth/oauth/:provider/callback", oauthHandler.HandleOAuthCallback)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authSvc))

	// Protected OAuth routes
	protected.GET("/auth/oauth/linked", oauthHandler.GetLinkedAccounts)
	protected.POST("/auth/oauth/:provider/link", oauthHandler.LinkOAuthAccount)
	protected.DELETE("/auth/oauth/:provider/unlink", oauthHandler.UnlinkOAuthAccount)

	protected.GET("/conversations", convHandler.GetConversations)
	protected.POST("/conversations", convHandler.CreateConversation) // Deprecated - for backward compatibility
	protected.GET("/conversations/:id", convHandler.GetConversation)
	protected.GET("/conversations/:id/messages", convHandler.GetMessages)

	// New message endpoint - handles both new conversations and existing ones
	protected.POST("/messages", convHandler.SendMessage)

	e.GET("/health", func(c echo.Context) error {
		if err := db.Health(c.Request().Context()); err != nil {
			return c.JSON(500, map[string]string{"status": "unhealthy", "error": err.Error()})
		}
		return c.JSON(200, map[string]string{"status": "healthy"})
	})

	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil {
			logger.WithError(err).Error("Server failed to start")
		}
	}()

	logger.WithField("port", cfg.Server.Port).Info("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	if err := e.Shutdown(context.TODO()); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}
}

// getEnvOrDefault gets environment variable with a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
