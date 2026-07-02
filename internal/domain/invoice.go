package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusUnpaid  InvoiceStatus = "unpaid"
	InvoiceStatusPartial InvoiceStatus = "partial"
	InvoiceStatusPaid    InvoiceStatus = "paid"
)

type Invoice struct {
	ID         uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkshopID uuid.UUID     `gorm:"type:uuid;not null;index" json:"workshop_id"`
	ServiceID  uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex" json:"service_id"`
	Service    Service       `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	InvoiceNo  string        `gorm:"uniqueIndex;not null" json:"invoice_no"`
	Subtotal   float64       `gorm:"type:numeric(12,2);not null" json:"subtotal"`
	Tax        float64       `gorm:"type:numeric(12,2);default:0" json:"tax"`
	Discount   float64       `gorm:"type:numeric(12,2);default:0" json:"discount"`
	Total      float64       `gorm:"type:numeric(12,2);not null" json:"total"`
	Status     InvoiceStatus `gorm:"type:varchar(20);default:'unpaid';index" json:"status"`
	DueDate    *time.Time    `json:"due_date"`
	Payments   []Payment     `gorm:"foreignKey:InvoiceID" json:"payments,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// DTOs
type CreateInvoiceRequest struct {
	ServiceID uuid.UUID `json:"service_id" validate:"required"`
	Tax       float64   `json:"tax" validate:"omitempty,min=0"`
	Discount  float64   `json:"discount" validate:"omitempty,min=0"`
	DueDate   string    `json:"due_date" validate:"omitempty"` // format: 2006-01-02
}

func (inv *Invoice) RecalculateStatus() {
	var paid float64
	for _, p := range inv.Payments {
		paid += p.Amount
	}
	switch {
	case paid <= 0:
		inv.Status = InvoiceStatusUnpaid
	case paid < inv.Total:
		inv.Status = InvoiceStatusPartial
	default:
		inv.Status = InvoiceStatusPaid
	}
}
