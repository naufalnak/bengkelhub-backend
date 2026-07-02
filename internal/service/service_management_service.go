package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type ServiceManagementService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateServiceRequest) (*domain.Service, error)
	GetAll(workshopID, ownerID uuid.UUID, status string, page, limit int) ([]domain.Service, int64, error)
	GetByID(id uuid.UUID) (*domain.Service, error)
	Update(id, ownerID uuid.UUID, req *domain.UpdateServiceRequest) (*domain.Service, error)
	Delete(id, ownerID uuid.UUID) error

	AddItem(serviceID, ownerID uuid.UUID, req *domain.AddServiceItemRequest) (*domain.ServiceItem, error)
	DeleteItem(itemID, serviceID, ownerID uuid.UUID) error
}

type serviceManagementService struct {
	serviceRepo  repository.ServiceRepository
	vehicleRepo  repository.VehicleRepository
	workshopRepo repository.WorkshopRepository
}

func NewServiceManagementService(serviceRepo repository.ServiceRepository, vehicleRepo repository.VehicleRepository, workshopRepo repository.WorkshopRepository) ServiceManagementService {
	return &serviceManagementService{serviceRepo, vehicleRepo, workshopRepo}
}

func generateServiceNo() string {
	return fmt.Sprintf("SRV-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))
}

func (s *serviceManagementService) Create(workshopID, ownerID uuid.UUID, req *domain.CreateServiceRequest) (*domain.Service, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	vehicle, err := s.vehicleRepo.FindByID(req.VehicleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found")
	}
	if err != nil {
		return nil, err
	}
	if vehicle.WorkshopID != workshopID {
		return nil, errors.New("vehicle does not belong to this workshop")
	}

	svc := &domain.Service{
		WorkshopID: workshopID,
		VehicleID:  req.VehicleID,
		MechanicID: req.MechanicID,
		ServiceNo:  generateServiceNo(),
		Complaint:  req.Complaint,
		Notes:      req.Notes,
		Status:     domain.ServiceStatusPending,
		StartDate:  time.Now(),
	}

	if err := s.serviceRepo.Create(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *serviceManagementService) GetAll(workshopID, ownerID uuid.UUID, status string, page, limit int) ([]domain.Service, int64, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.serviceRepo.FindByWorkshopID(workshopID, status, page, limit)
}

func (s *serviceManagementService) GetByID(id uuid.UUID) (*domain.Service, error) {
	svc, err := s.serviceRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("service not found")
	}
	return svc, err
}

func (s *serviceManagementService) Update(id, ownerID uuid.UUID, req *domain.UpdateServiceRequest) (*domain.Service, error) {
	svc, err := s.serviceRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, svc.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	if req.Diagnosis != "" {
		svc.Diagnosis = req.Diagnosis
	}
	if req.Notes != "" {
		svc.Notes = req.Notes
	}
	if req.MechanicID != nil {
		svc.MechanicID = req.MechanicID
	}
	if req.Status != "" {
		svc.Status = req.Status
		if req.Status == domain.ServiceStatusDone || req.Status == domain.ServiceStatusCancelled {
			now := time.Now()
			svc.EndDate = &now
		}
	}

	if err := s.serviceRepo.Update(svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *serviceManagementService) Delete(id, ownerID uuid.UUID) error {
	svc, err := s.serviceRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("service not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, svc.WorkshopID, ownerID); err != nil {
		return err
	}

	return s.serviceRepo.Delete(id)
}

func (s *serviceManagementService) AddItem(serviceID, ownerID uuid.UUID, req *domain.AddServiceItemRequest) (*domain.ServiceItem, error) {
	svc, err := s.serviceRepo.FindByID(serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, svc.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	qty, total := req.ComputeTotal()
	item := &domain.ServiceItem{
		ServiceID:   serviceID,
		Name:        req.Name,
		Description: req.Description,
		Qty:         qty,
		UnitPrice:   req.UnitPrice,
		Total:       total,
	}

	if err := s.serviceRepo.AddItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *serviceManagementService) DeleteItem(itemID, serviceID, ownerID uuid.UUID) error {
	svc, err := s.serviceRepo.FindByID(serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("service not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, svc.WorkshopID, ownerID); err != nil {
		return err
	}

	item, err := s.serviceRepo.FindItemByID(itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("service item not found")
	}
	if err != nil {
		return err
	}
	if item.ServiceID != serviceID {
		return errors.New("service item does not belong to this service")
	}

	return s.serviceRepo.DeleteItem(itemID)
}
