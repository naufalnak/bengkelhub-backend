package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/handler"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
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
 
	// ── Repositories ──────────────────────────────────────
	userRepo     := repository.NewUserRepository(config.DB)
	workshopRepo := repository.NewWorkshopRepository(config.DB)
 
	// ── Services ──────────────────────────────────────────
	authSvc     := service.NewAuthService(userRepo)
	workshopSvc := service.NewWorkshopService(workshopRepo)
 
	// ── Handlers ──────────────────────────────────────────
	authHandler     := handler.NewAuthHandler(authSvc)
	workshopHandler := handler.NewWorkshopHandler(workshopSvc)
 
	// ── Fiber app ─────────────────────────────────────────
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})
 
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))
 
	// ── Routes ────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
 
	v1 := app.Group("/api/v1")
 
	// Auth (public)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", middleware.Auth(), authHandler.Me)
 
	// Workshops
	workshops := v1.Group("/workshops")
	workshops.Get("/", workshopHandler.GetAll)                                                          // public
	workshops.Get("/my", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.GetMyWorkshops) // operator
	workshops.Get("/:id", workshopHandler.GetByID)                                                     // public
	workshops.Post("/", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Create)         // operator
	workshops.Patch("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Update)     // operator
	workshops.Delete("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Delete)    // operator
 
	// Start
	log.Printf("Server running on port %s", config.Cfg.AppPort)
	if err := app.Listen(":" + config.Cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
