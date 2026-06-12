package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
	"github.com/naufalnak/bengkelhub-backend/pkg/validator"
)

type SlotHandler struct {
	slotService  service.SlotService
	workshopRepo repository.WorkshopRepository
}

func NewSlotHandler(slotService service.SlotService, workshopRepo repository.WorkshopRepository) *SlotHandler {
	return &SlotHandler{slotService, workshopRepo}
}

// GET /api/v1/workshops/:workshopId/slots — public, list semua slot
// GET /api/v1/workshops/:workshopId/slots?available=true — hanya slot available
func (h *SlotHandler) GetByWorkshop(c *fiber.Ctx) error {
	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	onlyAvailable := c.Query("available") == "true"

	slots, err := h.slotService.GetByWorkshopID(workshopID, onlyAvailable)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch slots", nil)
	}

	// Tambahkan info remaining per slot
	type slotWithInfo struct {
		domain.Slot
		Remaining int  `json:"remaining"`
		Available bool `json:"available"`
	}

	result := make([]slotWithInfo, len(slots))
	for i, s := range slots {
		result[i] = slotWithInfo{
			Slot:      s,
			Remaining: s.RemainingSlots(),
			Available: s.IsAvailable(),
		}
	}

	return response.Success(c, fiber.StatusOK, "OK", result)
}

// GET /api/v1/slots/:id — public, detail slot
func (h *SlotHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid slot ID", nil)
	}

	slot, err := h.slotService.GetByID(id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"slot":      slot,
		"remaining": slot.RemainingSlots(),
		"available": slot.IsAvailable(),
	})
}

// POST /api/v1/workshops/:workshopId/slots — operator only
func (h *SlotHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	workshopID, err := uuid.Parse(c.Params("workshopId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.CreateSlotRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	slot, err := h.slotService.Create(workshopID, ownerID, &req)
	if err != nil {
		switch err.Error() {
		case "workshop not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusCreated, "Slot created", slot)
}

// PATCH /api/v1/slots/:id — operator only
func (h *SlotHandler) Update(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid slot ID", nil)
	}

	var req domain.UpdateSlotRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	slot, err := h.slotService.Update(id, ownerID, h.workshopRepo, &req)
	if err != nil {
		switch err.Error() {
		case "slot not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "Slot updated", slot)
}

// DELETE /api/v1/slots/:id — operator only
func (h *SlotHandler) Delete(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid slot ID", nil)
	}

	if err := h.slotService.Delete(id, ownerID, h.workshopRepo); err != nil {
		switch err.Error() {
		case "slot not found":
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		case "forbidden: you don't own this workshop":
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		case "cannot delete slot that already has bookings":
			return response.Error(c, fiber.StatusConflict, err.Error(), nil)
		default:
			return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
		}
	}

	return response.Success(c, fiber.StatusOK, "Slot deleted", nil)
}
