package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitConfig konfigurasi buat tiap tier rate limit
type RateLimitConfig struct {
	Max        int           // max request dalam window
	Expiration time.Duration // panjang window
	KeyPrefix  string        // prefix key di store
}

// globalConfig: semua endpoint, per IP
var globalConfig = RateLimitConfig{
	Max:        100,
	Expiration: 1 * time.Minute,
	KeyPrefix:  "rl:global:",
}

// authConfig: endpoint sensitif (login, register, resend), lebih ketat
var authConfig = RateLimitConfig{
	Max:        10,
	Expiration: 1 * time.Minute,
	KeyPrefix:  "rl:auth:",
}

// buildLimiter membuat Fiber middleware limiter dari config
func buildLimiter(cfg RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Gunakan X-Forwarded-For kalau ada (di belakang reverse proxy)
			ip := c.Get("X-Forwarded-For")
			if ip == "" {
				ip = c.IP()
			}
			return cfg.KeyPrefix + ip
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests, please slow down",
			})
		},
	})
}

// GlobalRateLimit: 100 req/menit per IP, berlaku untuk semua endpoint
func GlobalRateLimit() fiber.Handler {
	return buildLimiter(globalConfig)
}

// AuthRateLimit: 10 req/menit per IP, khusus endpoint auth yang sensitif
// (login, register, resend-verification)
func AuthRateLimit() fiber.Handler {
	return buildLimiter(authConfig)
}
