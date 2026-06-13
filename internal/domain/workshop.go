package domain

import (
	"time"

	"github.com/google/uuid"
)

type Workshop struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	Owner       User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Address     string    `gorm:"not null" json:"address"`
	Phone       string    `json:"phone"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Slot struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID uuid.UUID `gorm:"type:uuid;not null" json:"workshop_id"`
	Workshop   Workshop  `gorm:"foreignKey:WorkshopID" json:"workshop,omitempty"`
	Date       time.Time `gorm:"not null" json:"date"`
	MaxBooking int       `gorm:"default:5" json:"max_booking"`
	Booked     int       `gorm:"default:0" json:"booked"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// DTOs
type CreateWorkshopRequest struct {
	Name        string `json:"name" validate:"required,min=3"`
	Description string `json:"description"`
	Address     string `json:"address" validate:"required"`
	Phone       string `json:"phone" validate:"omitempty"`
}

type UpdateWorkshopRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3"`
	Description string `json:"description"`
	Address     string `json:"address" validate:"omitempty"`
	Phone       string `json:"phone" validate:"omitempty"`
	IsActive    *bool  `json:"is_active" validate:"omitempty"`
}

type CreateSlotRequest struct {
	Date       string `json:"date" validate:"required"`
	MaxBooking int    `json:"max_booking" validate:"omitempty,min=1"`
}

type UpdateSlotRequest struct {
	Date       string `json:"date" validate:"omitempty"`
	MaxBooking int    `json:"max_booking" validate:"omitempty,min=1"`
}

func (s *Slot) IsAvailable() bool {
	return s.Booked < s.MaxBooking
}

func (s *Slot) RemainingSlots() int {
	return s.MaxBooking - s.Booked
}