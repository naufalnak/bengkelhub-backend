package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"
)

type Payment struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID  uuid.UUID     `gorm:"type:uuid;not null;index" json:"workshop_id"`
	InvoiceID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"invoice_id"`
	Amount      float64       `gorm:"type:numeric(12,2);not null" json:"amount"`
	Method      PaymentMethod `gorm:"type:varchar(20);default:'cash'" json:"method"`
	ReferenceNo string        `json:"reference_no"`
	Notes       string        `json:"notes"`
	PaidAt      time.Time     `gorm:"default:now()" json:"paid_at"`
	CreatedAt   time.Time     `json:"created_at"`
}

// DTOs
type AddPaymentRequest struct {
	Amount      float64       `json:"amount" validate:"required,min=0.01"`
	Method      PaymentMethod `json:"method" validate:"omitempty,oneof=cash transfer qris"`
	ReferenceNo string        `json:"reference_no" validate:"omitempty"`
	Notes       string        `json:"notes" validate:"omitempty"`
}
