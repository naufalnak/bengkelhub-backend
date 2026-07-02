package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/middleware"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
	"github.com/naufalnak/bengkelhub-backend/pkg/validator"
	"net/url"
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

	return response.Success(c, fiber.StatusCreated, "Registration successful. Please check your email to verify your account.", result)
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
	userID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	user, err := h.authService.Me(userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "OK", user)
}

// GET /api/v1/auth/verify-email?token=xxx
// Dipanggil dari link di email → redirect ke frontend setelah proses
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.Error(c, fiber.StatusBadRequest, "Verification token is required", nil)
	}

	base := config.Cfg.AppBaseURL
	if err := h.authService.VerifyEmail(token); err != nil {
		// Redirect ke halaman error di frontend biar UX lebih smooth
		errURL := base + "/verify-email?status=error&message=" + url.QueryEscape(err.Error())
		return c.Redirect(errURL, fiber.StatusTemporaryRedirect)
	}

	return c.Redirect(base+"/verify-email?status=success", fiber.StatusTemporaryRedirect)
}

// POST /api/v1/auth/resend-verification
// Butuh JWT (user sudah login tapi belum verifikasi email)
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	userID := c.Locals(middleware.UserIDKey).(uuid.UUID)

	if err := h.authService.ResendVerification(userID); err != nil {
		return response.Error(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Success(c, fiber.StatusOK, "Verification email sent. Please check your inbox.", nil)
}
