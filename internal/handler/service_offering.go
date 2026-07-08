package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
	"github.com/naufalnak/bengkelhub-backend/pkg/validator"
)

type ServiceOfferingHandler struct {
	offeringService service.ServiceOfferingService
}

func NewServiceOfferingHandler(offeringService service.ServiceOfferingService) *ServiceOfferingHandler {
	return &ServiceOfferingHandler{offeringService}
}

// GET /api/v1/workshops/:workshopId/services-offered — publik, dipakai di
// halaman detail bengkel (customer) & halaman kelola layanan (operator)
func (h *ServiceOfferingHandler) GetAll(c *fiber.Ctx) error {
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	offerings, err := h.offeringService.GetByWorkshopID(workshopID)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "OK", offerings)
}

// POST /api/v1/workshops/:workshopId/services-offered — operator only
func (h *ServiceOfferingHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.CreateServiceOfferingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	offering, err := h.offeringService.Create(workshopID, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "Service offering created", offering)
}

// PATCH /api/v1/service-offerings/:id — operator only
func (h *ServiceOfferingHandler) Update(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID", nil)
	}

	var req domain.UpdateServiceOfferingRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	offering, err := h.offeringService.Update(id, ownerID, &req)
	if err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Service offering updated", offering)
}

// DELETE /api/v1/service-offerings/:id — operator only
func (h *ServiceOfferingHandler) Delete(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid ID", nil)
	}

	if err := h.offeringService.Delete(id, ownerID); err != nil {
		return handleOwnershipError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "Service offering deleted", nil)
}