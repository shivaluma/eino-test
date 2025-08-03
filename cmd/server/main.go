package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivaluma/eino-agent/config"
	"github.com/shivaluma/eino-agent/internal/auth"
	"github.com/shivaluma/eino-agent/internal/database"
	"github.com/shivaluma/eino-agent/internal/handlers"
	"github.com/shivaluma/eino-agent/internal/middleware"
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

	userRepo := repository.NewUserRepository(db)
	convRepo := repository.NewConversationRepository(db)
	authSvc := auth.NewService(cfg)

	authHandler := handlers.NewAuthHandler(userRepo, authSvc)
	convHandler := handlers.NewConversationHandler(convRepo, authSvc)

	e := echo.New()

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(middleware.CORSMiddleware())

	api := e.Group("/api/v1")

	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)
	api.POST("/token/refresh", authHandler.RefreshToken)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authSvc))

	protected.GET("/conversations", convHandler.GetConversations)
	protected.POST("/conversations", convHandler.CreateConversation)
	protected.GET("/conversations/:id", convHandler.GetConversation)
	protected.GET("/conversations/:id/messages", convHandler.GetMessages)

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