package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	Create(service *domain.Service) error
	FindByWorkshopID(workshopID uuid.UUID, status string, page, limit int) ([]domain.Service, int64, error)
	FindByID(id uuid.UUID) (*domain.Service, error)
	Update(service *domain.Service) error
	Delete(id uuid.UUID) error

	AddItem(item *domain.ServiceItem) error
	FindItemByID(id uuid.UUID) (*domain.ServiceItem, error)
	DeleteItem(id uuid.UUID) error
	SumItemsTotal(serviceID uuid.UUID) (float64, error)
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db}
}

func (r *serviceRepository) Create(service *domain.Service) error {
	return r.db.Create(service).Error
}

func (r *serviceRepository) FindByWorkshopID(workshopID uuid.UUID, status string, page, limit int) ([]domain.Service, int64, error) {
	var services []domain.Service
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Service{}).Where("workshop_id = ?", workshopID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Vehicle").Preload("Vehicle.Customer").Preload("Mechanic").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&services).Error

	return services, total, err
}

func (r *serviceRepository) FindByID(id uuid.UUID) (*domain.Service, error) {
	var service domain.Service
	err := r.db.
		Preload("Vehicle").Preload("Vehicle.Customer").Preload("Mechanic").Preload("ServiceItems").
		First(&service, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) Update(service *domain.Service) error {
	return r.db.Save(service).Error
}

func (r *serviceRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Service{}, "id = ?", id).Error
}

func (r *serviceRepository) AddItem(item *domain.ServiceItem) error {
	return r.db.Create(item).Error
}

func (r *serviceRepository) FindItemByID(id uuid.UUID) (*domain.ServiceItem, error) {
	var item domain.ServiceItem
	err := r.db.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *serviceRepository) DeleteItem(id uuid.UUID) error {
	return r.db.Delete(&domain.ServiceItem{}, "id = ?", id).Error
}

func (r *serviceRepository) SumItemsTotal(serviceID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.Model(&domain.ServiceItem{}).
		Where("service_id = ?", serviceID).
		Select("COALESCE(SUM(total), 0)").
		Scan(&total).Error
	return total, err
}
