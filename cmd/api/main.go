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
	"github.com/naufalnak/bengkelhub-backend/pkg/tasks"
)

func main() {
	config.Load()
	config.ConnectDB()
	tasks.InitClient()

	// AutoMigrate semua domain — field baru (MidtransOrderID, PaymentURL) otomatis ditambah
	if err := config.DB.AutoMigrate(
		&domain.User{},
		&domain.Workshop{},
		&domain.Slot{},
		&domain.Order{},
		&domain.Customer{},
		&domain.Vehicle{},
		&domain.Service{},
		&domain.ServiceItem{},
		&domain.Invoice{},
		&domain.Payment{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// ── Repositories ──────────────────────────────────────
	userRepo     := repository.NewUserRepository(config.DB)
	workshopRepo := repository.NewWorkshopRepository(config.DB)
	slotRepo     := repository.NewSlotRepository(config.DB)
	orderRepo    := repository.NewOrderRepository(config.DB)
	customerRepo := repository.NewCustomerRepository(config.DB)
	vehicleRepo  := repository.NewVehicleRepository(config.DB)
	serviceRepo  := repository.NewServiceRepository(config.DB)
	invoiceRepo  := repository.NewInvoiceRepository(config.DB)
	paymentRepo  := repository.NewPaymentRepository(config.DB)

	// ── Services ──────────────────────────────────────────
	authSvc        := service.NewAuthService(userRepo)
	workshopSvc    := service.NewWorkshopService(workshopRepo)
	slotSvc        := service.NewSlotService(slotRepo, workshopRepo)
	orderSvc       := service.NewOrderService(orderRepo, slotRepo, workshopRepo, userRepo)
	customerSvc    := service.NewCustomerService(customerRepo, workshopRepo)
	vehicleSvc     := service.NewVehicleService(vehicleRepo, customerRepo, workshopRepo)
	serviceMgmtSvc := service.NewServiceManagementService(serviceRepo, vehicleRepo, workshopRepo)
	invoiceSvc     := service.NewInvoiceService(invoiceRepo, paymentRepo, serviceRepo, workshopRepo)
	laporanSvc     := service.NewLaporanService(paymentRepo, workshopRepo)

	// ── Handlers ──────────────────────────────────────────
	authHandler     := handler.NewAuthHandler(authSvc)
	workshopHandler := handler.NewWorkshopHandler(workshopSvc)
	slotHandler     := handler.NewSlotHandler(slotSvc, workshopRepo)
	orderHandler    := handler.NewOrderHandler(orderSvc)
	customerHandler := handler.NewCustomerHandler(customerSvc)
	vehicleHandler  := handler.NewVehicleHandler(vehicleSvc)
	serviceHandler  := handler.NewServiceHandler(serviceMgmtSvc)
	invoiceHandler  := handler.NewInvoiceHandler(invoiceSvc)
	laporanHandler  := handler.NewLaporanHandler(laporanSvc)
	webhookHandler  := handler.NewWebhookHandler(invoiceSvc)

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
		AllowOrigins:     config.Cfg.CORSOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, ngrok-skip-browser-warning",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: false,
	}))
	app.Use(middleware.GlobalRateLimit())

	// ── Routes ────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	v1 := app.Group("/api/v1")

	// Auth
	auth := v1.Group("/auth")
	auth.Post("/register", middleware.AuthRateLimit(), authHandler.Register)
	auth.Post("/login", middleware.AuthRateLimit(), authHandler.Login)
	auth.Get("/me", middleware.Auth(), authHandler.Me)
	auth.Get("/verify-email", authHandler.VerifyEmail)
	auth.Post("/resend-verification", middleware.Auth(), middleware.AuthRateLimit(), authHandler.ResendVerification)

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
	workshops.Post("/:workshopId/slots/bulk", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), slotHandler.BulkCreate)
	workshops.Post("/:workshopId/slots", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), slotHandler.Create)

	// Orders (nested — operator)
	workshops.Get("/:workshopId/orders", middleware.Auth(), middleware.RequireRole(domain.RoleOperator), orderHandler.GetWorkshopOrders)

	// Internal management (operator only)
	opAuth := middleware.Auth()
	opRole := middleware.RequireRole(domain.RoleOperator)

	// Customers
	workshops.Get("/:workshopId/customers", opAuth, opRole, customerHandler.GetAll)
	workshops.Post("/:workshopId/customers", opAuth, opRole, customerHandler.Create)
	customers := v1.Group("/customers", opAuth, opRole)
	customers.Get("/:id", customerHandler.GetByID)
	customers.Patch("/:id", customerHandler.Update)
	customers.Delete("/:id", customerHandler.Delete)

	// Vehicles
	workshops.Get("/:workshopId/vehicles", opAuth, opRole, vehicleHandler.GetAll)
	workshops.Post("/:workshopId/vehicles", opAuth, opRole, vehicleHandler.Create)
	vehicles := v1.Group("/vehicles", opAuth, opRole)
	vehicles.Get("/:id", vehicleHandler.GetByID)
	vehicles.Patch("/:id", vehicleHandler.Update)
	vehicles.Delete("/:id", vehicleHandler.Delete)

	// Services
	workshops.Get("/:workshopId/services", opAuth, opRole, serviceHandler.GetAll)
	workshops.Post("/:workshopId/services", opAuth, opRole, serviceHandler.Create)
	services := v1.Group("/services", opAuth, opRole)
	services.Get("/:id", serviceHandler.GetByID)
	services.Patch("/:id", serviceHandler.Update)
	services.Delete("/:id", serviceHandler.Delete)
	services.Post("/:id/items", serviceHandler.AddItem)
	services.Delete("/:id/items/:itemId", serviceHandler.DeleteItem)

	// Invoices & Payments
	workshops.Get("/:workshopId/invoices", opAuth, opRole, invoiceHandler.GetAll)
	workshops.Post("/:workshopId/invoices", opAuth, opRole, invoiceHandler.Create)
	invoices := v1.Group("/invoices", opAuth, opRole)
	invoices.Get("/:id", invoiceHandler.GetByID)
	invoices.Post("/:id/payments", invoiceHandler.AddPayment)
	invoices.Delete("/:id/payments/:paymentId", invoiceHandler.DeletePayment)
	// Checkout: generate Midtrans payment URL
	invoices.Post("/:id/checkout", invoiceHandler.Checkout)

	// Laporan
	workshops.Get("/:workshopId/laporan", opAuth, opRole, laporanHandler.GetMonthly)

	// Slots standalone
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

	// Webhook Midtrans — PUBLIC, tidak butuh JWT, diverifikasi via signature
	v1.Post("/webhooks/payment", webhookHandler.HandlePayment)

	log.Printf("Server running on port %s", config.Cfg.AppPort)
	if err := app.Listen(":" + config.Cfg.AppPort); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
