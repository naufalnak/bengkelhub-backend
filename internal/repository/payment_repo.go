package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	FindByID(id uuid.UUID) (*domain.Payment, error)
	FindByInvoiceID(invoiceID uuid.UUID) ([]domain.Payment, error)
	Delete(id uuid.UUID) error

	// Laporan: total pemasukan per metode dalam rentang tanggal, milik 1 workshop
	SumByDateRange(workshopID uuid.UUID, start, end time.Time) (float64, error)
	FindByDateRange(workshopID uuid.UUID, start, end time.Time) ([]domain.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db}
}

func (r *paymentRepository) Create(payment *domain.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) FindByID(id uuid.UUID) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByInvoiceID(invoiceID uuid.UUID) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Where("invoice_id = ?", invoiceID).Order("paid_at DESC").Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Payment{}, "id = ?", id).Error
}

func (r *paymentRepository) SumByDateRange(workshopID uuid.UUID, start, end time.Time) (float64, error) {
	var total float64
	err := r.db.Model(&domain.Payment{}).
		Where("workshop_id = ? AND paid_at >= ? AND paid_at < ?", workshopID, start, end).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *paymentRepository) FindByDateRange(workshopID uuid.UUID, start, end time.Time) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.
		Where("workshop_id = ? AND paid_at >= ? AND paid_at < ?", workshopID, start, end).
		Order("paid_at DESC").
		Find(&payments).Error
	return payments, err
}
