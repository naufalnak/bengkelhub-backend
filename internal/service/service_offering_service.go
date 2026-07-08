package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type ServiceOfferingService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateServiceOfferingRequest) (*domain.ServiceOffering, error)
	// GetByWorkshopID sengaja publik (gak ada ownerID) — dipakai baik di
	// halaman kelola milik operator maupun halaman detail bengkel yang
	// customer lihat sebelum booking.
	GetByWorkshopID(workshopID uuid.UUID) ([]domain.ServiceOffering, error)
	Update(id, ownerID uuid.UUID, req *domain.UpdateServiceOfferingRequest) (*domain.ServiceOffering, error)
	Delete(id, ownerID uuid.UUID) error
}

type serviceOfferingService struct {
	offeringRepo repository.ServiceOfferingRepository
	workshopRepo repository.WorkshopRepository
}

func NewServiceOfferingService(
	offeringRepo repository.ServiceOfferingRepository,
	workshopRepo repository.WorkshopRepository,
) ServiceOfferingService {
	return &serviceOfferingService{offeringRepo, workshopRepo}
}

func (s *serviceOfferingService) Create(workshopID, ownerID uuid.UUID, req *domain.CreateServiceOfferingRequest) (*domain.ServiceOffering, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	offering := &domain.ServiceOffering{
		WorkshopID:     workshopID,
		Name:           req.Name,
		Description:    req.Description,
		EstimatedPrice: req.EstimatedPrice,
	}
	if err := s.offeringRepo.Create(offering); err != nil {
		return nil, err
	}
	return offering, nil
}

func (s *serviceOfferingService) GetByWorkshopID(workshopID uuid.UUID) ([]domain.ServiceOffering, error) {
	return s.offeringRepo.FindByWorkshopID(workshopID)
}

func (s *serviceOfferingService) Update(id, ownerID uuid.UUID, req *domain.UpdateServiceOfferingRequest) (*domain.ServiceOffering, error) {
	offering, err := s.offeringRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("service offering not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, offering.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	if req.Name != "" {
		offering.Name = req.Name
	}
	if req.Description != "" {
		offering.Description = req.Description
	}
	if req.EstimatedPrice != nil {
		offering.EstimatedPrice = req.EstimatedPrice
	}

	if err := s.offeringRepo.Update(offering); err != nil {
		return nil, err
	}
	return offering, nil
}

func (s *serviceOfferingService) Delete(id, ownerID uuid.UUID) error {
	offering, err := s.offeringRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("service offering not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, offering.WorkshopID, ownerID); err != nil {
		return err
	}

	return s.offeringRepo.Delete(id)
}