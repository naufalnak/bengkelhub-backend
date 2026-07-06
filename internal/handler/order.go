package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
	"github.com/naufalnak/bengkelhub-backend/pkg/validator"
	"strconv"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService}
}

// POST /api/v1/orders — customer only, buat booking
func (h *OrderHandler) Create(c *fiber.Ctx) error {
	customerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	var req domain.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	order, err := h.orderService.Create(customerID, &req)
	if err != nil {
		switch err.Error() {
		case "workshop not found", "slot not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "slot is full or already passed":
			return response.Error(c, fiber.StatusConflict, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusCreated, "Booking created", order)
}

// GET /api/v1/orders/:id — customer (own) atau operator (workshop-nya)
func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	requesterID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	requesterRole := c.Locals(middleware.UserRoleKey).(domain.Role)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", nil)
	}

	order, err := h.orderService.GetByID(id, requesterID, requesterRole)
	if err != nil {
		if err.Error() == "order not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "forbidden" {
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", order)
}

// GET /api/v1/orders/my — customer, list order sendiri
func (h *OrderHandler) GetMyOrders(c *fiber.Ctx) error {
	customerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	orders, total, err := h.orderService.GetMyOrders(customerID, page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch orders", nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  orders,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/workshops/:workshopId/orders — operator, list order di workshop-nya
func (h *OrderHandler) GetWorkshopOrders(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	orders, total, err := h.orderService.GetWorkshopOrders(workshopID, ownerID, page, limit)
	if err != nil {
		if err.Error() == "workshop not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "forbidden: you don't own this workshop" {
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  orders,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// PATCH /api/v1/orders/:id/status — operator only, update status
func (h *OrderHandler) UpdateStatus(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", nil)
	}

	var req domain.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	order, err := h.orderService.UpdateStatus(id, ownerID, &req)
	if err != nil {
		switch err.Error() {
		case "order not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "Status updated", order)
}

// PATCH /api/v1/orders/:id/cancel — customer only, cancel order sendiri
func (h *OrderHandler) Cancel(c *fiber.Ctx) error {
	customerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", nil)
	}

	if err := h.orderService.Cancel(id, customerID); err != nil {
		switch err.Error() {
		case "order not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "forbidden":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "Order cancelled", nil)
}

// POST /api/v1/orders/:id/convert-to-service — operator only
// Konversi booking (Order) jadi Customer + Vehicle + Service internal sekaligus,
// biar operator gak perlu input manual data yang sama berkali-kali.
func (h *OrderHandler) ConvertToService(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid order ID", nil)
	}

	svc, err := h.orderService.ConvertToService(id, ownerID)
	if err != nil {
		switch err.Error() {
		case "order not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "order already converted to service":
			return response.Error(c, fiber.StatusConflict, err.Error(), nil)
		case "cannot convert a cancelled order", "customer phone not available":
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusCreated, "Order converted to service", svc)
}