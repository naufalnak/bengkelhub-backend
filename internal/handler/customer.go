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

type CustomerHandler struct {
	customerService service.CustomerService
}

func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService}
}

func handleOwnershipError(c *fiber.Ctx, err error) error {
	switch err.Error() {
	case "workshop not found", "customer not found", "vehicle not found", "service not found", "invoice not found", "payment not found", "service item not found":
		return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
	case "forbidden: you don't own this workshop":
		return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
	default:
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}
}

// GET /api/v1/workshops/:workshopId/customers
func (h *CustomerHandler) GetAll(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	search := c.Query("search", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	customers, total, err := h.customerService.GetAll(workshopID, ownerID, search, page, limit)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  customers,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/customers/:id
func (h *CustomerHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid customer ID", nil)
	}

	customer, err := h.customerService.GetByID(id)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", customer)
}

// POST /api/v1/workshops/:workshopId/customers
func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	customer, err := h.customerService.Create(workshopID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Customer created", customer)
}

// PATCH /api/v1/customers/:id
func (h *CustomerHandler) Update(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid customer ID", nil)
	}

	var req domain.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	customer, err := h.customerService.Update(id, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Customer updated", customer)
}

// DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid customer ID", nil)
	}

	if err := h.customerService.Delete(id, ownerID); err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Customer deleted", nil)
}
