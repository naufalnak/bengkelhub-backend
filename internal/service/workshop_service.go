package service

import (
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/pkg/geo"
	"gorm.io/gorm"
)

type WorkshopService interface {
	Create(ownerID uuid.UUID, req *domain.CreateWorkshopRequest) (*domain.Workshop, error)
	// GetAll: lat & lng opsional (nil kalau customer gak share lokasi/GPS mati).
	// Kalau diisi, hasil diurutkan dari yang PALING DEKAT, dan tiap workshop
	// dapat field DistanceKM terisi.
	GetAll(page, limit int, lat, lng *float64) ([]domain.Workshop, int64, error)
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
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsActive:    true,
	}

	if err := s.workshopRepo.Create(workshop); err != nil {
		return nil, err
	}

	return workshop, nil
}

func (s *workshopService) GetAll(page, limit int, lat, lng *float64) ([]domain.Workshop, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Gak ada lokasi customer → jalur biasa, paginate langsung dari database
	// (lebih murah, gak perlu tarik semua row).
	if lat == nil || lng == nil {
		return s.workshopRepo.FindAll(page, limit)
	}

	// Ada lokasi customer → butuh tau jarak SEMUA workshop aktif dulu, baru bisa
	// diurutkan dari yang terdekat sebelum di-paginate. Untuk skala BengkelHub
	// sekarang ini masih murah; kalau nanti workshop-nya udah ribuan, baru worth
	// pindah ke query jarak native di database (mis. PostGIS).
	all, err := s.workshopRepo.FindAllActive()
	if err != nil {
		return nil, 0, err
	}

	for i := range all {
		if all[i].Latitude != nil && all[i].Longitude != nil {
			d := geo.DistanceKM(*lat, *lng, *all[i].Latitude, *all[i].Longitude)
			all[i].DistanceKM = &d
		}
	}

	// Urutkan dari yang terdekat. Workshop yang belum diisi lokasinya ditaruh
	// paling belakang (bukan dianggap "0 km" atau ikut dianggap ambigu).
	sort.SliceStable(all, func(i, j int) bool {
		di, dj := all[i].DistanceKM, all[j].DistanceKM
		if di == nil {
			return false
		}
		if dj == nil {
			return true
		}
		return *di < *dj
	})

	total := int64(len(all))
	start := (page - 1) * limit
	if start > len(all) {
		start = len(all)
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}

	return all[start:end], total, nil
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
	if req.QrisImageURL != "" {
		workshop.QrisImageURL = req.QrisImageURL
	}
	if req.Latitude != nil {
		workshop.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		workshop.Longitude = req.Longitude
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