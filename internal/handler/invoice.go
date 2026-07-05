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

// POST /api/v1/invoices/:id/checkout — operator only
// Generate Midtrans payment URL, kembalikan invoice dengan payment_url terisi
func (h *InvoiceHandler) Checkout(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid invoice ID", nil)
	}

	invoice, err := h.invoiceService.Checkout(id, ownerID)
	if err != nil {
		switch err.Error() {
		case "invoice not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "invoice already paid", "invoice already fully paid":
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "Payment URL generated", fiber.Map{
		"payment_url":       invoice.PaymentURL,
		"midtrans_order_id": invoice.MidtransOrderID,
		"invoice":           invoice,
	})
}

// POST /api/v1/invoices/:id/send-whatsapp — operator only
// Kirim ringkasan invoice (no. invoice, total, status, link bayar) ke WA customer via Fonnte.
func (h *InvoiceHandler) SendWhatsapp(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid invoice ID", nil)
	}

	if err := h.invoiceService.SendWhatsapp(id, ownerID); err != nil {
		switch err.Error() {
		case "invoice not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "customer phone not available":
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadGateway, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "WhatsApp message sent", nil)
}