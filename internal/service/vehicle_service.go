package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type VehicleService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateVehicleRequest) (*domain.Vehicle, error)
	GetAll(workshopID, ownerID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error)
	GetByID(id uuid.UUID) (*domain.Vehicle, error)
	Update(id, ownerID uuid.UUID, req *domain.UpdateVehicleRequest) (*domain.Vehicle, error)
	Delete(id, ownerID uuid.UUID) error
}

type vehicleService struct {
	vehicleRepo  repository.VehicleRepository
	customerRepo repository.CustomerRepository
	workshopRepo repository.WorkshopRepository
}

func NewVehicleService(vehicleRepo repository.VehicleRepository, customerRepo repository.CustomerRepository, workshopRepo repository.WorkshopRepository) VehicleService {
	return &vehicleService{vehicleRepo, customerRepo, workshopRepo}
}

func (s *vehicleService) Create(workshopID, ownerID uuid.UUID, req *domain.CreateVehicleRequest) (*domain.Vehicle, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	customer, err := s.customerRepo.FindByID(req.CustomerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("customer not found")
	}
	if err != nil {
		return nil, err
	}
	if customer.WorkshopID != workshopID {
		return nil, errors.New("customer does not belong to this workshop")
	}

	vehicle := &domain.Vehicle{
		WorkshopID:  workshopID,
		CustomerID:  req.CustomerID,
		PlateNumber: req.PlateNumber,
		Brand:       req.Brand,
		Model:       req.Model,
		Year:        req.Year,
		Color:       req.Color,
		EngineCC:    req.EngineCC,
	}

	if err := s.vehicleRepo.Create(vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) GetAll(workshopID, ownerID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.vehicleRepo.FindByWorkshopID(workshopID, search, page, limit)
}

func (s *vehicleService) GetByID(id uuid.UUID) (*domain.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found")
	}
	return vehicle, err
}

func (s *vehicleService) Update(id, ownerID uuid.UUID, req *domain.UpdateVehicleRequest) (*domain.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, vehicle.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	if req.PlateNumber != "" {
		vehicle.PlateNumber = req.PlateNumber
	}
	if req.Brand != "" {
		vehicle.Brand = req.Brand
	}
	if req.Model != "" {
		vehicle.Model = req.Model
	}
	if req.Year != 0 {
		vehicle.Year = req.Year
	}
	if req.Color != "" {
		vehicle.Color = req.Color
	}
	if req.EngineCC != 0 {
		vehicle.EngineCC = req.EngineCC
	}

	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Delete(id, ownerID uuid.UUID) error {
	vehicle, err := s.vehicleRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("vehicle not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, vehicle.WorkshopID, ownerID); err != nil {
		return err
	}

	return s.vehicleRepo.Delete(id)
}
