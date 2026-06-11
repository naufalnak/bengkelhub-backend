package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/naufalnak/bengkelku-api/config"
	"github.com/naufalnak/bengkelku-api/internal/domain"
	"github.com/naufalnak/bengkelku-api/internal/handler"
	"github.com/naufalnak/bengkelku-api/internal/repository"
	"github.com/naufalnak/bengkelku-api/internal/service"
	"github.com/naufalnak/bengkelku-api/pkg/middleware"
)

func main() {
	// Load config
	config.Load()

	// Connect DB
	config.ConnectDB()

	// Auto migrate
	if err := config.DB.AutoMigrate(
		&domain.User{},
		&domain.Workshop{},
		&domain.Slot{},
		&domain.Order{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// Init layers
	userRepo := repository.NewUserRepository(config.DB)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	// Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", middleware.Auth(), authHandler.Me)

	// Start server
	log.Printf("Server running on port %s", config.Cfg.AppPort)
	if err := app.Listen(":" + config.Cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
