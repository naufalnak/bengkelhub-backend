package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type CustomerService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateCustomerRequest) (*domain.Customer, error)
	GetAll(workshopID, ownerID uuid.UUID, search string, page, limit int) ([]domain.Customer, int64, error)
	GetByID(id uuid.UUID) (*domain.Customer, error)
	Update(id, ownerID uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.Customer, error)
	Delete(id, ownerID uuid.UUID) error
}

type customerService struct {
	customerRepo repository.CustomerRepository
	workshopRepo repository.WorkshopRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository, workshopRepo repository.WorkshopRepository) CustomerService {
	return &customerService{customerRepo, workshopRepo}
}

// verifyWorkshopOwner memastikan ownerID adalah pemilik workshopID
func verifyWorkshopOwner(workshopRepo repository.WorkshopRepository, workshopID, ownerID uuid.UUID) (*domain.Workshop, error) {
	workshop, err := workshopRepo.FindByID(workshopID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("workshop not found")
	}
	if err != nil {
		return nil, err
	}
	if workshop.OwnerID != ownerID {
		return nil, errors.New("forbidden: you don't own this workshop")
	}
	return workshop, nil
}

func (s *customerService) Create(workshopID, ownerID uuid.UUID, req *domain.CreateCustomerRequest) (*domain.Customer, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	customer := &domain.Customer{
		WorkshopID: workshopID,
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		Address:    req.Address,
	}

	if err := s.customerRepo.Create(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) GetAll(workshopID, ownerID uuid.UUID, search string, page, limit int) ([]domain.Customer, int64, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.customerRepo.FindByWorkshopID(workshopID, search, page, limit)
}

func (s *customerService) GetByID(id uuid.UUID) (*domain.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("customer not found")
	}
	return customer, err
}

func (s *customerService) Update(id, ownerID uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("customer not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, customer.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}
	if req.Address != "" {
		customer.Address = req.Address
	}

	if err := s.customerRepo.Update(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) Delete(id, ownerID uuid.UUID) error {
	customer, err := s.customerRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("customer not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, customer.WorkshopID, ownerID); err != nil {
		return err
	}

	return s.customerRepo.Delete(id)
}
