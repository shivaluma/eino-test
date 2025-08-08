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

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations on startup
	log.Println("Running database migrations...")
	migrator := migrations.NewMigrator(db.Pool, "migrations", cfg)
	if err := migrator.Migrate(context.Background()); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

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
		log.Fatalf("Failed to get AI provider: %v", err)
	}

	model, err := provider.CreateChatModel(ctx)
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	aiService := ai.NewService(model, &ai.Config{
		DefaultProvider: provider.GetName(),
	})

	authHandler := handlers.NewAuthHandler(userRepo, authSvc)
	oauthHandler := handlers.NewOAuthHandler(userRepo, oauthRepo, authSvc, oauthSvc, cfg.OAuth.FrontendURL)
	convHandler := handlers.NewConversationHandler(convRepo, authSvc, aiService)

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(echomiddleware.Logger())
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
			log.Printf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := e.Shutdown(context.TODO()); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
