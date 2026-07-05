package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	Create(invoice *domain.Invoice) error
	FindByWorkshopID(workshopID uuid.UUID, status string, page, limit int) ([]domain.Invoice, int64, error)
	FindByID(id uuid.UUID) (*domain.Invoice, error)
	FindByServiceID(serviceID uuid.UUID) (*domain.Invoice, error)
	FindByMidtransOrderID(orderID string) (*domain.Invoice, error)
	Update(invoice *domain.Invoice) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db}
}

func (r *invoiceRepository) Create(invoice *domain.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) FindByWorkshopID(workshopID uuid.UUID, status string, page, limit int) ([]domain.Invoice, int64, error) {
	var invoices []domain.Invoice
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Invoice{}).Where("workshop_id = ?", workshopID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Service").Preload("Service.Vehicle").
		Preload("Service.Vehicle.Customer").Preload("Payments").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&invoices).Error

	return invoices, total, err
}

func (r *invoiceRepository) FindByID(id uuid.UUID) (*domain.Invoice, error) {
	var invoice domain.Invoice
	err := r.db.
		Preload("Service").Preload("Service.Vehicle").
		Preload("Service.Vehicle.Customer").Preload("Payments").
		First(&invoice, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByServiceID(serviceID uuid.UUID) (*domain.Invoice, error) {
	var invoice domain.Invoice
	err := r.db.Where("service_id = ?", serviceID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) FindByMidtransOrderID(orderID string) (*domain.Invoice, error) {
	var invoice domain.Invoice
	err := r.db.
		Preload("Payments").
		Where("midtrans_order_id = ?", orderID).First(&invoice).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *invoiceRepository) Update(invoice *domain.Invoice) error {
	return r.db.Save(invoice).Error
}
