package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type SlotRepository interface {
	Create(slot *domain.Slot) error
	FindByWorkshopID(workshopID uuid.UUID) ([]domain.Slot, error)
	FindAvailableByWorkshopID(workshopID uuid.UUID) ([]domain.Slot, error)
	FindByID(id uuid.UUID) (*domain.Slot, error)
	Update(slot *domain.Slot) error
	Delete(id uuid.UUID) error
	IncrementBooked(id uuid.UUID) error
	DecrementBooked(id uuid.UUID) error
}

type slotRepository struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) SlotRepository {
	return &slotRepository{db}
}

func (r *slotRepository) Create(slot *domain.Slot) error {
	return r.db.Create(slot).Error
}

func (r *slotRepository) FindByWorkshopID(workshopID uuid.UUID) ([]domain.Slot, error) {
	var slots []domain.Slot
	err := r.db.
		Where("workshop_id = ?", workshopID).
		Order("date ASC").
		Find(&slots).Error
	return slots, err
}

func (r *slotRepository) FindAvailableByWorkshopID(workshopID uuid.UUID) ([]domain.Slot, error) {
	var slots []domain.Slot
	err := r.db.
		Where("workshop_id = ? AND booked < max_booking AND date > NOW()", workshopID).
		Order("date ASC").
		Find(&slots).Error
	return slots, err
}

func (r *slotRepository) FindByID(id uuid.UUID) (*domain.Slot, error) {
	var slot domain.Slot
	err := r.db.First(&slot, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *slotRepository) Update(slot *domain.Slot) error {
	return r.db.Save(slot).Error
}

func (r *slotRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Slot{}, "id = ?", id).Error
}

func (r *slotRepository) IncrementBooked(id uuid.UUID) error {
	return r.db.Model(&domain.Slot{}).
		Where("id = ?", id).
		UpdateColumn("booked", gorm.Expr("booked + 1")).Error
}

func (r *slotRepository) DecrementBooked(id uuid.UUID) error {
	return r.db.Model(&domain.Slot{}).
		Where("id = ? AND booked > 0", id).
		UpdateColumn("booked", gorm.Expr("booked - 1")).Error
}
