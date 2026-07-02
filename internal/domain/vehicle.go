package domain

import (
	"time"

	"github.com/google/uuid"
)

type Vehicle struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID  uuid.UUID `gorm:"type:uuid;not null;index" json:"workshop_id"`
	CustomerID  uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Customer    Customer  `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	PlateNumber string    `gorm:"not null" json:"plate_number"`
	Brand       string    `gorm:"not null" json:"brand"`
	Model       string    `gorm:"not null" json:"model"`
	Year        int       `json:"year"`
	Color       string    `json:"color"`
	EngineCC    int       `json:"engine_cc"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DTOs
type CreateVehicleRequest struct {
	CustomerID  uuid.UUID `json:"customer_id" validate:"required"`
	PlateNumber string    `json:"plate_number" validate:"required"`
	Brand       string    `json:"brand" validate:"required"`
	Model       string    `json:"model" validate:"required"`
	Year        int       `json:"year" validate:"omitempty,min=1900"`
	Color       string    `json:"color" validate:"omitempty"`
	EngineCC    int       `json:"engine_cc" validate:"omitempty,min=0"`
}

type UpdateVehicleRequest struct {
	PlateNumber string `json:"plate_number" validate:"omitempty"`
	Brand       string `json:"brand" validate:"omitempty"`
	Model       string `json:"model" validate:"omitempty"`
	Year        int    `json:"year" validate:"omitempty,min=1900"`
	Color       string `json:"color" validate:"omitempty"`
	EngineCC    int    `json:"engine_cc" validate:"omitempty,min=0"`
}
