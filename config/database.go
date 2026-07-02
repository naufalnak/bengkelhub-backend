package config

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	gormConfig := &gorm.Config{}
	if Cfg.AppEnv == "development" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	// PreferSimpleProtocol = true wajib kalau DATABASE_URL mengarah ke
	// Supabase Connection Pooler (port 6543 / PgBouncer transaction mode),
	// karena PgBouncer transaction mode tidak mendukung prepared statement.
	// Untuk direct connection (port 5432) ini tetap aman dipakai.
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  Cfg.DatabaseURL,
		PreferSimpleProtocol: true,
	}), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
}