package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusDone      BookingStatus = "done"
	BookingStatusCancelled BookingStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CustomerID  uuid.UUID     `gorm:"type:uuid;not null" json:"customer_id"`
	Customer    User          `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	WorkshopID  uuid.UUID     `gorm:"type:uuid;not null" json:"workshop_id"`
	Workshop    Workshop      `gorm:"foreignKey:WorkshopID" json:"workshop,omitempty"`
	SlotID      uuid.UUID     `gorm:"type:uuid;not null" json:"slot_id"`
	Slot        Slot          `gorm:"foreignKey:SlotID" json:"slot,omitempty"`
	Status      BookingStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Notes       string        `json:"notes"`
	VehicleType string        `gorm:"not null" json:"vehicle_type"`
	VehiclePlate string       `gorm:"not null" json:"vehicle_plate"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// DTOs
type CreateOrderRequest struct {
	WorkshopID   string `json:"workshop_id" validate:"required,uuid"`
	SlotID       string `json:"slot_id" validate:"required,uuid"`
	VehicleType  string `json:"vehicle_type" validate:"required"`
	VehiclePlate string `json:"vehicle_plate" validate:"required"`
	Notes        string `json:"notes"`
}

type UpdateOrderStatusRequest struct {
	Status BookingStatus `json:"status" validate:"required,oneof=confirmed done cancelled"`
}
