package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"gorm.io/gorm"
)

type CreateSlotRequest struct {
    Date       string `json:"date" validate:"required"`
    MaxBooking int    `json:"max_booking" validate:"omitempty,min=1"`
}

type UpdateSlotRequest struct {
    Date       string `json:"date" validate:"omitempty"`
    MaxBooking int    `json:"max_booking" validate:"omitempty,min=1"`
}

type SlotService interface {
	Create(workshopID uuid.UUID, ownerID uuid.UUID, req *domain.CreateSlotRequest) (*domain.Slot, error)
	GetByWorkshopID(workshopID uuid.UUID, onlyAvailable bool) ([]domain.Slot, error)
	GetByID(id uuid.UUID) (*domain.Slot, error)
	Update(id uuid.UUID, ownerID uuid.UUID, workshopRepo repository.WorkshopRepository, req *domain.UpdateSlotRequest) (*domain.Slot, error)
	Delete(id uuid.UUID, ownerID uuid.UUID, workshopRepo repository.WorkshopRepository) error
}

type slotService struct {
	slotRepo     repository.SlotRepository
	workshopRepo repository.WorkshopRepository
}

func NewSlotService(slotRepo repository.SlotRepository, workshopRepo repository.WorkshopRepository) SlotService {
	return &slotService{slotRepo, workshopRepo}
}

func (s *slotService) Create(workshopID uuid.UUID, ownerID uuid.UUID, req *domain.CreateSlotRequest) (*domain.Slot, error) {
	// Verify workshop exists and belongs to owner
	workshop, err := s.workshopRepo.FindByID(workshopID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("workshop not found")
	}
	if err != nil {
		return nil, err
	}
	if workshop.OwnerID != ownerID {
		return nil, errors.New("forbidden: you don't own this workshop")
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use RFC3339 (e.g. 2025-01-15T09:00:00Z)")
	}
	if date.Before(time.Now()) {
		return nil, errors.New("slot date must be in the future")
	}

	maxBooking := req.MaxBooking
	if maxBooking == 0 {
		maxBooking = 5
	}

	slot := &domain.Slot{
		WorkshopID: workshopID,
		Date:       date,
		MaxBooking: maxBooking,
		Booked:     0,
	}

	if err := s.slotRepo.Create(slot); err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *slotService) GetByWorkshopID(workshopID uuid.UUID, onlyAvailable bool) ([]domain.Slot, error) {
	if onlyAvailable {
		return s.slotRepo.FindAvailableByWorkshopID(workshopID)
	}
	return s.slotRepo.FindByWorkshopID(workshopID)
}

func (s *slotService) GetByID(id uuid.UUID) (*domain.Slot, error) {
	slot, err := s.slotRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("slot not found")
	}
	return slot, err
}

func (s *slotService) Update(id uuid.UUID, ownerID uuid.UUID, workshopRepo repository.WorkshopRepository, req *domain.UpdateSlotRequest) (*domain.Slot, error) {
	slot, err := s.slotRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("slot not found")
	}
	if err != nil {
		return nil, err
	}

	// Verify ownership via workshop
	workshop, err := s.workshopRepo.FindByID(slot.WorkshopID)
	if err != nil {
		return nil, err
	}
	if workshop.OwnerID != ownerID {
		return nil, errors.New("forbidden: you don't own this workshop")
	}

	if req.Date != "" {
		date, err := time.Parse(time.RFC3339, req.Date)
		if err != nil {
			return nil, errors.New("invalid date format, use RFC3339")
		}
		slot.Date = date
	}
	if req.MaxBooking > 0 {
		if req.MaxBooking < slot.Booked {
			return nil, errors.New("max_booking cannot be less than current booked count")
		}
		slot.MaxBooking = req.MaxBooking
	}

	if err := s.slotRepo.Update(slot); err != nil {
		return nil, err
	}

	return slot, nil
}

func (s *slotService) Delete(id uuid.UUID, ownerID uuid.UUID, workshopRepo repository.WorkshopRepository) error {
	slot, err := s.slotRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("slot not found")
	}
	if err != nil {
		return err
	}

	if slot.Booked > 0 {
		return errors.New("cannot delete slot that already has bookings")
	}

	workshop, err := s.workshopRepo.FindByID(slot.WorkshopID)
	if err != nil {
		return err
	}
	if workshop.OwnerID != ownerID {
		return errors.New("forbidden: you don't own this workshop")
	}

	return s.slotRepo.Delete(id)
}
