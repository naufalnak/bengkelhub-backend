package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	AppPort           string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiresIn      string
	RedisURL          string
	FonnteToken       string
	ResendAPIKey      string
	AppBaseURL        string
	CORSOrigins       string
	MidtransServerKey string
	MidtransClientKey string
	MidtransEnv       string // "sandbox" atau "production"
}

var Cfg *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	Cfg = &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTSecret:         getEnv("JWT_SECRET", "secret"),
		JWTExpiresIn:      getEnv("JWT_EXPIRES_IN", "24h"),
		RedisURL:          getEnv("REDIS_URL", ""),
		FonnteToken:       getEnv("FONNTE_TOKEN", ""),
		ResendAPIKey:      getEnv("RESEND_API_KEY", ""),
		AppBaseURL:        getEnv("APP_BASE_URL", "http://localhost:3000"),
		CORSOrigins:       getEnv("APP_CORS_ORIGINS", "http://localhost:3000"),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransClientKey: getEnv("MIDTRANS_CLIENT_KEY", ""),
		MidtransEnv:       getEnv("MIDTRANS_ENV", "sandbox"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}