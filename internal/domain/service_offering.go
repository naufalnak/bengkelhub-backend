package domain

import (
	"time"

	"github.com/google/uuid"
)

// ServiceOffering adalah daftar layanan yang bengkel TAWARKAN ke customer
// (semacam menu, mis. "Ganti Oli", "Servis AC", "Tune Up") — beda dengan
// Service yang merupakan pekerjaan servis AKTUAL untuk kendaraan tertentu.
// Ditampilkan di halaman detail bengkel (publik) biar customer tahu bisa
// ngapain aja di bengkel itu sebelum booking.
type ServiceOffering struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID  uuid.UUID `gorm:"type:uuid;not null;index" json:"workshop_id"`
	Workshop    Workshop  `gorm:"foreignKey:WorkshopID" json:"workshop,omitempty"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	// EstimatedPrice opsional — kasih gambaran kisaran harga ke customer.
	// BUKAN harga final; harga pasti tetap ditentukan pas invoice dibuat.
	EstimatedPrice *float64  `json:"estimated_price"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DTOs
type CreateServiceOfferingRequest struct {
	Name           string   `json:"name" validate:"required,min=2"`
	Description    string   `json:"description"`
	EstimatedPrice *float64 `json:"estimated_price" validate:"omitempty,gte=0"`
}

type UpdateServiceOfferingRequest struct {
	Name           string   `json:"name" validate:"omitempty,min=2"`
	Description    string   `json:"description"`
	EstimatedPrice *float64 `json:"estimated_price" validate:"omitempty,gte=0"`
}