package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
	"github.com/naufalnak/bengkelhub-backend/pkg/validator"
)

type InvoiceHandler struct {
	invoiceService service.InvoiceService
}

func NewInvoiceHandler(invoiceService service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService}
}

// GET /api/v1/workshops/:workshopId/invoices
func (h *InvoiceHandler) GetAll(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	invoices, total, err := h.invoiceService.GetAll(workshopID, ownerID, status, page, limit)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  invoices,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid invoice ID", nil)
	}

	invoice, err := h.invoiceService.GetByID(id)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", invoice)
}

// POST /api/v1/workshops/:workshopId/invoices
func (h *InvoiceHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.CreateInvoiceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	invoice, err := h.invoiceService.Create(workshopID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Invoice created", invoice)
}

// POST /api/v1/invoices/:id/payments
func (h *InvoiceHandler) AddPayment(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid invoice ID", nil)
	}

	var req domain.AddPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	payment, err := h.invoiceService.AddPayment(invoiceID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Payment added", payment)
}

// DELETE /api/v1/invoices/:id/payments/:paymentId
func (h *InvoiceHandler) DeletePayment(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid invoice ID", nil)
	}
	paymentID, err := uuid.Parse(c.Params("paymentId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid payment ID", nil)
	}

	if err := h.invoiceService.DeletePayment(paymentID, invoiceID, ownerID); err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Payment deleted", nil)
}
