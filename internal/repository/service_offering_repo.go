package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type ServiceOfferingRepository interface {
	Create(offering *domain.ServiceOffering) error
	FindByWorkshopID(workshopID uuid.UUID) ([]domain.ServiceOffering, error)
	FindByID(id uuid.UUID) (*domain.ServiceOffering, error)
	Update(offering *domain.ServiceOffering) error
	Delete(id uuid.UUID) error
}

type serviceOfferingRepository struct {
	db *gorm.DB
}

func NewServiceOfferingRepository(db *gorm.DB) ServiceOfferingRepository {
	return &serviceOfferingRepository{db}
}

func (r *serviceOfferingRepository) Create(offering *domain.ServiceOffering) error {
	return r.db.Create(offering).Error
}

// FindByWorkshopID gak dipaginasi sengaja — daftar layanan biasanya cuma
// belasan item, ditampilin semua sekaligus baik di halaman kelola (operator)
// maupun halaman detail bengkel (publik).
func (r *serviceOfferingRepository) FindByWorkshopID(workshopID uuid.UUID) ([]domain.ServiceOffering, error) {
	var offerings []domain.ServiceOffering
	err := r.db.Where("workshop_id = ?", workshopID).
		Order("created_at ASC").
		Find(&offerings).Error
	return offerings, err
}

func (r *serviceOfferingRepository) FindByID(id uuid.UUID) (*domain.ServiceOffering, error) {
	var offering domain.ServiceOffering
	err := r.db.First(&offering, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &offering, nil
}

func (r *serviceOfferingRepository) Update(offering *domain.ServiceOffering) error {
	return r.db.Save(offering).Error
}

func (r *serviceOfferingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.ServiceOffering{}, "id = ?", id).Error
}