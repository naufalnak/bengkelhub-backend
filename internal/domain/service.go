package domain

import (
	"time"

	"github.com/google/uuid"
)

type ServiceStatus string

const (
	ServiceStatusPending    ServiceStatus = "pending"
	ServiceStatusInProgress ServiceStatus = "in_progress"
	ServiceStatusDone       ServiceStatus = "done"
	ServiceStatusCancelled  ServiceStatus = "cancelled"
)

type Service struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"workshop_id"`
	VehicleID    uuid.UUID     `gorm:"type:uuid;not null;index" json:"vehicle_id"`
	Vehicle      Vehicle       `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	MechanicID   *uuid.UUID    `gorm:"type:uuid;index" json:"mechanic_id"`
	Mechanic     *User         `gorm:"foreignKey:MechanicID" json:"mechanic,omitempty"`
	ServiceNo    string        `gorm:"uniqueIndex;not null" json:"service_no"`
	Complaint    string        `gorm:"not null" json:"complaint"`
	Diagnosis    string        `json:"diagnosis"`
	Notes        string        `json:"notes"`
	Status       ServiceStatus `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	StartDate    time.Time     `gorm:"default:now()" json:"start_date"`
	EndDate      *time.Time    `json:"end_date"`
	ServiceItems []ServiceItem `gorm:"foreignKey:ServiceID" json:"service_items,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type ServiceItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ServiceID   uuid.UUID `gorm:"type:uuid;not null;index" json:"service_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Qty         int       `gorm:"default:1" json:"qty"`
	UnitPrice   float64   `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Total       float64   `gorm:"type:numeric(12,2);not null" json:"total"`
	CreatedAt   time.Time `json:"created_at"`
}

// DTOs
type CreateServiceRequest struct {
	VehicleID  uuid.UUID  `json:"vehicle_id" validate:"required"`
	MechanicID *uuid.UUID `json:"mechanic_id" validate:"omitempty"`
	Complaint  string     `json:"complaint" validate:"required"`
	Notes      string     `json:"notes" validate:"omitempty"`
}

type UpdateServiceRequest struct {
	Diagnosis  string        `json:"diagnosis" validate:"omitempty"`
	Notes      string        `json:"notes" validate:"omitempty"`
	MechanicID *uuid.UUID    `json:"mechanic_id" validate:"omitempty"`
	Status     ServiceStatus `json:"status" validate:"omitempty,oneof=pending in_progress done cancelled"`
}

type AddServiceItemRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"omitempty"`
	Qty         int     `json:"qty" validate:"omitempty,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,min=0"`
}

func (i *AddServiceItemRequest) ComputeTotal() (qty int, total float64) {
	qty = i.Qty
	if qty < 1 {
		qty = 1
	}
	total = float64(qty) * i.UnitPrice
	return
}
