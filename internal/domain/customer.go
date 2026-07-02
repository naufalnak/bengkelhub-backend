package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID uuid.UUID `gorm:"type:uuid;not null;index" json:"workshop_id"`
	Workshop   Workshop  `gorm:"foreignKey:WorkshopID" json:"workshop,omitempty"`
	Name       string    `gorm:"not null" json:"name"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DTOs
type CreateCustomerRequest struct {
	Name    string `json:"name" validate:"required,min=2"`
	Phone   string `json:"phone" validate:"omitempty,e164"`
	Email   string `json:"email" validate:"omitempty,email"`
	Address string `json:"address" validate:"omitempty"`
}

type UpdateCustomerRequest struct {
	Name    string `json:"name" validate:"omitempty,min=2"`
	Phone   string `json:"phone" validate:"omitempty,e164"`
	Email   string `json:"email" validate:"omitempty,email"`
	Address string `json:"address" validate:"omitempty"`
}
