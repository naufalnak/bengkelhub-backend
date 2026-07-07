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

type WorkshopHandler struct {
	workshopService service.WorkshopService
}

func NewWorkshopHandler(workshopService service.WorkshopService) *WorkshopHandler {
	return &WorkshopHandler{workshopService}
}

// GET /api/v1/workshops — public, list semua workshop aktif
// Query opsional ?lat=&lng= — kalau diisi (browser share GPS customer), hasil
// diurutkan dari yang PALING DEKAT & tiap workshop dapat field distance_km.
func (h *WorkshopHandler) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	var lat, lng *float64
	if v, err := strconv.ParseFloat(c.Query("lat"), 64); err == nil {
		lat = &v
	}
	if v, err := strconv.ParseFloat(c.Query("lng"), 64); err == nil {
		lng = &v
	}

	workshops, total, err := h.workshopService.GetAll(page, limit, lat, lng)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch workshops", nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  workshops,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GET /api/v1/workshops/:id — public, detail workshop
func (h *WorkshopHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	workshop, err := h.workshopService.GetByID(id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", workshop)
}

// GET /api/v1/workshops/my — operator, list workshop milik sendiri
func (h *WorkshopHandler) GetMyWorkshops(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	workshops, total, err := h.workshopService.GetMyWorkshops(ownerID, page, limit)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to fetch workshops", nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{
		"data":  workshops,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// POST /api/v1/workshops — operator only
func (h *WorkshopHandler) Create(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	var req domain.CreateWorkshopRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	workshop, err := h.workshopService.Create(ownerID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Workshop created", workshop)
}

// PATCH /api/v1/workshops/:id — operator only, owner only
func (h *WorkshopHandler) Update(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	var req domain.UpdateWorkshopRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	workshop, err := h.workshopService.Update(id, ownerID, &req)
	if err != nil {
		if err.Error() == "workshop not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "forbidden: you don't own this workshop" {
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Workshop updated", workshop)
}

// DELETE /api/v1/workshops/:id — operator only, owner only
func (h *WorkshopHandler) Delete(c *fiber.Ctx) error {
	ownerID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid workshop ID", nil)
	}

	if err := h.workshopService.Delete(id, ownerID); err != nil {
		if err.Error() == "workshop not found" {
			return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
		}
		if err.Error() == "forbidden: you don't own this workshop" {
			return response.Error(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return response.Error(c, fiber.StatusInternalServerError, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Workshop deleted", nil)
}