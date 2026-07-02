package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
)

type LaporanData struct {
	Month        int              `json:"month"`
	Year         int              `json:"year"`
	TotalIncome  float64          `json:"total_income"`
	Transactions []domain.Payment `json:"transactions"`
}

type LaporanService interface {
	GetMonthly(workshopID, ownerID uuid.UUID, month, year int) (*LaporanData, error)
}

type laporanService struct {
	paymentRepo  repository.PaymentRepository
	workshopRepo repository.WorkshopRepository
}

func NewLaporanService(paymentRepo repository.PaymentRepository, workshopRepo repository.WorkshopRepository) LaporanService {
	return &laporanService{paymentRepo, workshopRepo}
}

func (s *laporanService) GetMonthly(workshopID, ownerID uuid.UUID, month, year int) (*LaporanData, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}
	if year < 2000 {
		year = time.Now().Year()
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	total, err := s.paymentRepo.SumByDateRange(workshopID, start, end)
	if err != nil {
		return nil, err
	}

	transactions, err := s.paymentRepo.FindByDateRange(workshopID, start, end)
	if err != nil {
		return nil, err
	}

	return &LaporanData{
		Month:        month,
		Year:         year,
		TotalIncome:  total,
		Transactions: transactions,
	}, nil
}
