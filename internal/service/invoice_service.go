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

type InvoiceService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateInvoiceRequest) (*domain.Invoice, error)
	GetAll(workshopID, ownerID uuid.UUID, status string, page, limit int) ([]domain.Invoice, int64, error)
	GetByID(id uuid.UUID) (*domain.Invoice, error)

	AddPayment(invoiceID, ownerID uuid.UUID, req *domain.AddPaymentRequest) (*domain.Payment, error)
	DeletePayment(paymentID, invoiceID, ownerID uuid.UUID) error
}

type invoiceService struct {
	invoiceRepo  repository.InvoiceRepository
	paymentRepo  repository.PaymentRepository
	serviceRepo  repository.ServiceRepository
	workshopRepo repository.WorkshopRepository
}

func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
	serviceRepo repository.ServiceRepository,
	workshopRepo repository.WorkshopRepository,
) InvoiceService {
	return &invoiceService{invoiceRepo, paymentRepo, serviceRepo, workshopRepo}
}

func generateInvoiceNo() string {
	return fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))
}

func (s *invoiceService) Create(workshopID, ownerID uuid.UUID, req *domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, err
	}

	svc, err := s.serviceRepo.FindByID(req.ServiceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, err
	}
	if svc.WorkshopID != workshopID {
		return nil, errors.New("service does not belong to this workshop")
	}

	if _, err := s.invoiceRepo.FindByServiceID(req.ServiceID); err == nil {
		return nil, errors.New("invoice already exists for this service")
	}

	subtotal, err := s.serviceRepo.SumItemsTotal(req.ServiceID)
	if err != nil {
		return nil, err
	}

	total := subtotal + req.Tax - req.Discount
	if total < 0 {
		total = 0
	}

	var dueDate *time.Time
	if req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &parsed
	}

	invoice := &domain.Invoice{
		WorkshopID: workshopID,
		ServiceID:  req.ServiceID,
		InvoiceNo:  generateInvoiceNo(),
		Subtotal:   subtotal,
		Tax:        req.Tax,
		Discount:   req.Discount,
		Total:      total,
		Status:     domain.InvoiceStatusUnpaid,
		DueDate:    dueDate,
	}

	if err := s.invoiceRepo.Create(invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *invoiceService) GetAll(workshopID, ownerID uuid.UUID, status string, page, limit int) ([]domain.Invoice, int64, error) {
	if _, err := verifyWorkshopOwner(s.workshopRepo, workshopID, ownerID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.invoiceRepo.FindByWorkshopID(workshopID, status, page, limit)
}

func (s *invoiceService) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	invoice, err := s.invoiceRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invoice not found")
	}
	return invoice, err
}

func (s *invoiceService) AddPayment(invoiceID, ownerID uuid.UUID, req *domain.AddPaymentRequest) (*domain.Payment, error) {
	invoice, err := s.invoiceRepo.FindByID(invoiceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invoice not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, invoice.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	method := req.Method
	if method == "" {
		method = domain.PaymentMethodCash
	}

	payment := &domain.Payment{
		WorkshopID:  invoice.WorkshopID,
		InvoiceID:   invoiceID,
		Amount:      req.Amount,
		Method:      method,
		ReferenceNo: req.ReferenceNo,
		Notes:       req.Notes,
		PaidAt:      time.Now(),
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	// Recalculate invoice status berdasarkan total payment terbaru
	payments, err := s.paymentRepo.FindByInvoiceID(invoiceID)
	if err != nil {
		return nil, err
	}
	invoice.Payments = payments
	invoice.RecalculateStatus()
	if err := s.invoiceRepo.Update(invoice); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *invoiceService) DeletePayment(paymentID, invoiceID, ownerID uuid.UUID) error {
	invoice, err := s.invoiceRepo.FindByID(invoiceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("invoice not found")
	}
	if err != nil {
		return err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, invoice.WorkshopID, ownerID); err != nil {
		return err
	}

	payment, err := s.paymentRepo.FindByID(paymentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("payment not found")
	}
	if err != nil {
		return err
	}
	if payment.InvoiceID != invoiceID {
		return errors.New("payment does not belong to this invoice")
	}

	if err := s.paymentRepo.Delete(paymentID); err != nil {
		return err
	}

	payments, err := s.paymentRepo.FindByInvoiceID(invoiceID)
	if err != nil {
		return err
	}
	invoice.Payments = payments
	invoice.RecalculateStatus()
	return s.invoiceRepo.Update(invoice)
}
