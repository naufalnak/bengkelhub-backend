package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *domain.Order) error
	FindByID(id uuid.UUID) (*domain.Order, error)
	FindByIDWithRelations(id uuid.UUID) (*domain.Order, error)
	FindByCustomerID(customerID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	FindByWorkshopID(workshopID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	UpdateStatus(id uuid.UUID, status domain.BookingStatus) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) Create(order *domain.Order) error {
	if err := r.db.Create(order).Error; err != nil {
		return err
	}
	// Reload dengan relasi setelah create
	return r.db.
		Preload("Customer").
		Preload("Workshop").
		Preload("Slot").
		First(order, "id = ?", order.ID).Error
}

func (r *orderRepository) FindByID(id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByIDWithRelations(id uuid.UUID) (*domain.Order, error) {
	var order domain.Order
	err := r.db.
		Preload("Customer").
		Preload("Workshop").
		Preload("Slot").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByCustomerID(customerID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&domain.Order{}).Where("customer_id = ?", customerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Where("customer_id = ?", customerID).
		Preload("Workshop").
		Preload("Slot").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) FindByWorkshopID(workshopID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	var orders []domain.Order
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&domain.Order{}).Where("workshop_id = ?", workshopID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Where("workshop_id = ?", workshopID).
		Preload("Customer").
		Preload("Slot").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) UpdateStatus(id uuid.UUID, status domain.BookingStatus) error {
	return r.db.Model(&domain.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}