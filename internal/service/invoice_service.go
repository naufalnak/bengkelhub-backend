package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/pkg/fonnte"
	mt "github.com/naufalnak/bengkelhub-backend/pkg/midtrans"
	"gorm.io/gorm"
)

type InvoiceService interface {
	Create(workshopID, ownerID uuid.UUID, req *domain.CreateInvoiceRequest) (*domain.Invoice, error)
	GetAll(workshopID, ownerID uuid.UUID, status string, page, limit int) ([]domain.Invoice, int64, error)
	GetByID(id uuid.UUID) (*domain.Invoice, error)
	AddPayment(invoiceID, ownerID uuid.UUID, req *domain.AddPaymentRequest) (*domain.Payment, error)
	DeletePayment(paymentID, invoiceID, ownerID uuid.UUID) error
	Checkout(invoiceID, ownerID uuid.UUID) (*domain.Invoice, error)
	SendWhatsapp(invoiceID, ownerID uuid.UUID) error
	HandleWebhook(payload mt.WebhookPayload) error
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

func (s *invoiceService) Checkout(invoiceID, ownerID uuid.UUID) (*domain.Invoice, error) {
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

	if invoice.Status == domain.InvoiceStatusPaid {
		return nil, errors.New("invoice already paid")
	}

	var totalPaid float64
	for _, p := range invoice.Payments {
		totalPaid += p.Amount
	}
	remaining := invoice.Total - totalPaid
	if remaining <= 0 {
		return nil, errors.New("invoice already fully paid")
	}

	midtransOrderID := fmt.Sprintf("%s-%d", invoice.InvoiceNo, time.Now().Unix())

	customerDetail := mt.CustomerDetail{FirstName: "Pelanggan"}
	if invoice.Service.Vehicle.Customer.Name != "" {
		customerDetail.FirstName = invoice.Service.Vehicle.Customer.Name
		customerDetail.Phone = invoice.Service.Vehicle.Customer.Phone
		customerDetail.Email = invoice.Service.Vehicle.Customer.Email
	}

	req := mt.CreateTransactionRequest{
		TransactionDetail: mt.TransactionDetail{
			OrderID:     midtransOrderID,
			GrossAmount: remaining,
		},
		CustomerDetail: customerDetail,
		ItemDetails: []mt.ItemDetail{
			{
				ID:       invoice.InvoiceNo,
				Name:     fmt.Sprintf("Invoice %s - %s", invoice.InvoiceNo, invoice.Service.Vehicle.PlateNumber),
				Price:    remaining,
				Quantity: 1,
			},
		},
		Callbacks: &mt.Callbacks{
			Finish: config.Cfg.AppBaseURL + "/payment/finish?invoice_id=" + invoiceID.String(),
		},
	}

	result, err := mt.CreateTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans transaction: %w", err)
	}

	invoice.MidtransOrderID = midtransOrderID
	invoice.PaymentURL = result.RedirectURL
	if err := s.invoiceRepo.Update(invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}

func (s *invoiceService) SendWhatsapp(invoiceID, ownerID uuid.UUID) error {
	invoice, err := s.invoiceRepo.FindByID(invoiceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("invoice not found")
	}
	if err != nil {
		return err
	}

	workshop, err := verifyWorkshopOwner(s.workshopRepo, invoice.WorkshopID, ownerID)
	if err != nil {
		return err
	}

	customer := invoice.Service.Vehicle.Customer
	if customer.Phone == "" {
		return errors.New("customer phone not available")
	}

	msg := buildInvoiceWhatsappMessage(invoice, workshop.Name, customer.Name)

	if err := fonnte.Send(customer.Phone, msg); err != nil {
		return fmt.Errorf("failed to send WA message: %w", err)
	}

	return nil
}

func (s *invoiceService) HandleWebhook(payload mt.WebhookPayload) error {
	if !mt.VerifySignature(payload) {
		return errors.New("invalid webhook signature")
	}

	invoice, err := s.invoiceRepo.FindByMidtransOrderID(payload.OrderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("invoice not found for order_id: " + payload.OrderID)
	}
	if err != nil {
		return err
	}

	if invoice.Status == domain.InvoiceStatusPaid {
		return nil
	}

	if mt.IsPaymentSuccess(payload) {
		var amount float64
		fmt.Sscanf(payload.GrossAmount, "%f", &amount)

		payment := &domain.Payment{
			WorkshopID:  invoice.WorkshopID,
			InvoiceID:   invoice.ID,
			Amount:      amount,
			Method:      domain.PaymentMethod(mt.PaymentMethodFromType(payload.PaymentType)),
			ReferenceNo: payload.TransactionID,
			Notes:       "via Midtrans - " + payload.PaymentType,
			PaidAt:      time.Now(),
		}

		if err := s.paymentRepo.Create(payment); err != nil {
			return fmt.Errorf("failed to create payment record: %w", err)
		}

		payments, err := s.paymentRepo.FindByInvoiceID(invoice.ID)
		if err != nil {
			return err
		}
		invoice.Payments = payments
		invoice.RecalculateStatus()
		if err := s.invoiceRepo.Update(invoice); err != nil {
			return err
		}
	}

	return nil
}