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
	config.Load()
	config.ConnectDB()

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
	slotRepo     := repository.NewSlotRepository(config.DB)
	orderRepo    := repository.NewOrderRepository(config.DB)

	// ── Services ──────────────────────────────────────────
	authSvc     := service.NewAuthService(userRepo)
	workshopSvc := service.NewWorkshopService(workshopRepo)
	slotSvc     := service.NewSlotService(slotRepo, workshopRepo)
	orderSvc    := service.NewOrderService(orderRepo, slotRepo, workshopRepo)

	// ── Handlers ──────────────────────────────────────────
	authHandler     := handler.NewAuthHandler(authSvc)
	workshopHandler := handler.NewWorkshopHandler(workshopSvc)
	slotHandler     := handler.NewSlotHandler(slotSvc, workshopRepo)
	orderHandler    := handler.NewOrderHandler(orderSvc)

	// ── Fiber ─────────────────────────────────────────────
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

	// Auth
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Get("/me", middleware.Auth(), authHandler.Me)

	// Workshops
	workshops := v1.Group("/workshops")
	workshops.Get("/", workshopHandler.GetAll)
	workshops.Get("/my", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.GetMyWorkshops)
	workshops.Get("/:id", workshopHandler.GetByID)
	workshops.Post("/", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Create)
	workshops.Patch("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Update)
	workshops.Delete("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), workshopHandler.Delete)

	// Slots (nested)
	workshops.Get("/:workshopId/slots", slotHandler.GetByWorkshop)
	workshops.Post("/:workshopId/slots", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), slotHandler.Create)

	// Orders (nested under workshop — operator)
	workshops.Get("/:workshopId/orders", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), orderHandler.GetWorkshopOrders)

	// Slots (standalone)
	slots := v1.Group("/slots")
	slots.Get("/:id", slotHandler.GetByID)
	slots.Patch("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), slotHandler.Update)
	slots.Delete("/:id", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), slotHandler.Delete)

	// Orders
	orders := v1.Group("/orders", middleware.Auth())
	orders.Post("/", middleware.RequireRole(domain.RoleCustomer), orderHandler.Create)
	orders.Get("/my", middleware.RequireRole(domain.RoleCustomer), orderHandler.GetMyOrders)
	orders.Get("/:id", orderHandler.GetByID)
	orders.Patch("/:id/status", middleware.RequireRole(domain.RoleOperator), orderHandler.UpdateStatus)
	orders.Patch("/:id/cancel", middleware.RequireRole(domain.RoleCustomer), orderHandler.Cancel)

	log.Printf("Server running on port %s", config.Cfg.AppPort)
	if err := app.Listen(":" + config.Cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}