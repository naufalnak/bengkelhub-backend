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

type ServiceHandler struct {
	serviceManagementService service.ServiceManagementService
}

func NewServiceHandler(serviceManagementService service.ServiceManagementService) *ServiceHandler {
	return &ServiceHandler{serviceManagementService}
}

// GET /api/v1/workshops/:workshopId/services
func (h *ServiceHandler) GetAll(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	services, total, err := h.serviceManagementService.GetAll(workshopID, ownerID, status, page, limit)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  services,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/services/:id
func (h *ServiceHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid service ID", nil)
	}

	svc, err := h.serviceManagementService.GetByID(id)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", svc)
}

// POST /api/v1/workshops/:workshopId/services
func (h *ServiceHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	svc, err := h.serviceManagementService.Create(workshopID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Service created", svc)
}

// PATCH /api/v1/services/:id
func (h *ServiceHandler) Update(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid service ID", nil)
	}

	var req domain.UpdateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	svc, err := h.serviceManagementService.Update(id, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Service updated", svc)
}

// DELETE /api/v1/services/:id
func (h *ServiceHandler) Delete(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid service ID", nil)
	}

	if err := h.serviceManagementService.Delete(id, ownerID); err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Service deleted", nil)
}

// POST /api/v1/services/:id/items
func (h *ServiceHandler) AddItem(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid service ID", nil)
	}

	var req domain.AddServiceItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	item, err := h.serviceManagementService.AddItem(serviceID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Service item added", item)
}

// DELETE /api/v1/services/:id/items/:itemId
func (h *ServiceHandler) DeleteItem(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	serviceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid service ID", nil)
	}
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid item ID", nil)
	}

	if err := h.serviceManagementService.DeleteItem(itemID, serviceID, ownerID); err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Service item deleted", nil)
}
