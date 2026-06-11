package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/naufalnak/bengkelku-api/internal/domain"
	"github.com/naufalnak/bengkelku-api/internal/service"
	"github.com/naufalnak/bengkelku-api/pkg/middleware"
	"github.com/naufalnak/bengkelku-api/pkg/response"
	"github.com/naufalnak/bengkelku-api/pkg/validator"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		return response.Error(c, fiber.StatusConflict, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusCreated, "Registration successful", result)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := validator.Validate(&req); errs != nil {
		return response.ValidationError(c, errs)
	}

	result, err := h.authService.Login(&req)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Login successful", result)
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals(middleware.UserIDKey)
	return response.Success(c, fiber.StatusOK, "OK", fiber.Map{"user_id": userID})
}
