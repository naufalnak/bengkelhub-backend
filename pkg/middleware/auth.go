package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/naufalnak/bengkelku-api/internal/domain"
	"github.com/naufalnak/bengkelku-api/pkg/jwt"
	"github.com/naufalnak/bengkelku-api/pkg/response"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"

func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return response.Error(c, fiber.StatusUnauthorized, "Missing or invalid token", nil)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.Parse(tokenStr)
		if err != nil {
			return response.Error(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)
		}

		c.Locals(UserIDKey, claims.UserID)
		c.Locals(UserRoleKey, claims.Role)
		return c.Next()
	}
}

func RequireRole(roles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals(UserRoleKey).(domain.Role)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, "Forbidden: insufficient permissions", nil)
	}
}
