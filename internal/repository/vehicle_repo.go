package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(vehicle *domain.Vehicle) error
	FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error)
	FindByCustomerID(customerID uuid.UUID) ([]domain.Vehicle, error)
	FindByID(id uuid.UUID) (*domain.Vehicle, error)
	Update(vehicle *domain.Vehicle) error
	Delete(id uuid.UUID) error
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db}
}

func (r *vehicleRepository) Create(vehicle *domain.Vehicle) error {
	return r.db.Create(vehicle).Error
}

func (r *vehicleRepository) FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error) {
	var vehicles []domain.Vehicle
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Vehicle{}).Where("workshop_id = ?", workshopID)

	if search != "" {
		query = query.Where("plate_number ILIKE ? OR brand ILIKE ? OR model ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Customer").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&vehicles).Error

	return vehicles, total, err
}

func (r *vehicleRepository) FindByCustomerID(customerID uuid.UUID) ([]domain.Vehicle, error) {
	var vehicles []domain.Vehicle
	err := r.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindByID(id uuid.UUID) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.Preload("Customer").First(&vehicle, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) Update(vehicle *domain.Vehicle) error {
	return r.db.Save(vehicle).Error
}

func (r *vehicleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Vehicle{}, "id = ?", id).Error
}
