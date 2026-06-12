package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type WorkshopRepository interface {
	Create(workshop *domain.Workshop) error
	FindAll(page, limit int) ([]domain.Workshop, int64, error)
	FindByID(id uuid.UUID) (*domain.Workshop, error)
	FindByOwnerID(ownerID uuid.UUID, page, limit int) ([]domain.Workshop, int64, error)
	Update(workshop *domain.Workshop) error
	Delete(id uuid.UUID) error
}

type workshopRepository struct {
	db *gorm.DB
}

func NewWorkshopRepository(db *gorm.DB) WorkshopRepository {
	return &workshopRepository{db}
}

func (r *workshopRepository) Create(workshop *domain.Workshop) error {
	return r.db.Create(workshop).Error
}

func (r *workshopRepository) FindAll(page, limit int) ([]domain.Workshop, int64, error) {
	var workshops []domain.Workshop
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&domain.Workshop{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Where("is_active = ?", true).
		Preload("Owner").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&workshops).Error

	return workshops, total, err
}

func (r *workshopRepository) FindByID(id uuid.UUID) (*domain.Workshop, error) {
	var workshop domain.Workshop
	err := r.db.Preload("Owner").First(&workshop, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workshop, nil
}

func (r *workshopRepository) FindByOwnerID(ownerID uuid.UUID, page, limit int) ([]domain.Workshop, int64, error) {
	var workshops []domain.Workshop
	var total int64

	offset := (page - 1) * limit

	if err := r.db.Model(&domain.Workshop{}).Where("owner_id = ?", ownerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Where("owner_id = ?", ownerID).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&workshops).Error

	return workshops, total, err
}

func (r *workshopRepository) Update(workshop *domain.Workshop) error {
	return r.db.Save(workshop).Error
}

func (r *workshopRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Workshop{}, "id = ?", id).Error
}
