package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	DatabaseURL  string
	JWTSecret    string
	JWTExpiresIn string
	RedisURL     string
	FonnteToken  string
	ResendAPIKey string
	AppBaseURL   string // dipakai buat generate link verifikasi email, contoh: https://bengkelhub.vercel.app
	CORSOrigins  string // comma-separated allowed origins, contoh: https://bengkelhub.vercel.app,http://localhost:3000
}

var Cfg *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	Cfg = &Config{
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("APP_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		JWTSecret:    getEnv("JWT_SECRET", "secret"),
		JWTExpiresIn: getEnv("JWT_EXPIRES_IN", "24h"),
		RedisURL:     getEnv("REDIS_URL", ""),
		FonnteToken:  getEnv("FONNTE_TOKEN", ""),
		ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:3000"),
		CORSOrigins:  getEnv("APP_CORS_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
