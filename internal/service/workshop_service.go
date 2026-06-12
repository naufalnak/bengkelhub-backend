package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type WorkshopService interface {
	Create(ownerID uuid.UUID, req *domain.CreateWorkshopRequest) (*domain.Workshop, error)
	GetAll(page, limit int) ([]domain.Workshop, int64, error)
	GetByID(id uuid.UUID) (*domain.Workshop, error)
	GetMyWorkshops(ownerID uuid.UUID, page, limit int) ([]domain.Workshop, int64, error)
	Update(id uuid.UUID, ownerID uuid.UUID, req *domain.UpdateWorkshopRequest) (*domain.Workshop, error)
	Delete(id uuid.UUID, ownerID uuid.UUID) error
}

type workshopService struct {
	workshopRepo repository.WorkshopRepository
}

func NewWorkshopService(workshopRepo repository.WorkshopRepository) WorkshopService {
	return &workshopService{workshopRepo}
}

func (s *workshopService) Create(ownerID uuid.UUID, req *domain.CreateWorkshopRequest) (*domain.Workshop, error) {
	workshop := &domain.Workshop{
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		IsActive:    true,
	}

	if err := s.workshopRepo.Create(workshop); err != nil {
		return nil, err
	}

	return workshop, nil
}

func (s *workshopService) GetAll(page, limit int) ([]domain.Workshop, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.workshopRepo.FindAll(page, limit)
}

func (s *workshopService) GetByID(id uuid.UUID) (*domain.Workshop, error) {
	workshop, err := s.workshopRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("workshop not found")
	}
	return workshop, err
}

func (s *workshopService) GetMyWorkshops(ownerID uuid.UUID, page, limit int) ([]domain.Workshop, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.workshopRepo.FindByOwnerID(ownerID, page, limit)
}

func (s *workshopService) Update(id uuid.UUID, ownerID uuid.UUID, req *domain.UpdateWorkshopRequest) (*domain.Workshop, error) {
	workshop, err := s.workshopRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("workshop not found")
	}
	if err != nil {
		return nil, err
	}

	// Only owner can update
	if workshop.OwnerID != ownerID {
		return nil, errors.New("forbidden: you don't own this workshop")
	}

	if req.Name != "" {
		workshop.Name = req.Name
	}
	if req.Description != "" {
		workshop.Description = req.Description
	}
	if req.Address != "" {
		workshop.Address = req.Address
	}
	if req.Phone != "" {
		workshop.Phone = req.Phone
	}
	if req.IsActive != nil {
		workshop.IsActive = *req.IsActive
	}

	if err := s.workshopRepo.Update(workshop); err != nil {
		return nil, err
	}

	return workshop, nil
}

func (s *workshopService) Delete(id uuid.UUID, ownerID uuid.UUID) error {
	workshop, err := s.workshopRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("workshop not found")
	}
	if err != nil {
		return err
	}

	if workshop.OwnerID != ownerID {
		return errors.New("forbidden: you don't own this workshop")
	}

	return s.workshopRepo.Delete(id)
}
