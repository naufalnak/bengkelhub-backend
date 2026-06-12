package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/pkg/fonnte"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(customerID uuid.UUID, req *domain.CreateOrderRequest) (*domain.Order, error)
	GetByID(id uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role) (*domain.Order, error)
	GetMyOrders(customerID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	GetWorkshopOrders(workshopID uuid.UUID, ownerID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	UpdateStatus(id uuid.UUID, ownerID uuid.UUID, req *domain.UpdateOrderStatusRequest) (*domain.Order, error)
	Cancel(id uuid.UUID, customerID uuid.UUID) error
}

type orderService struct {
	orderRepo    repository.OrderRepository
	slotRepo     repository.SlotRepository
	workshopRepo repository.WorkshopRepository
	userRepo     repository.UserRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	slotRepo repository.SlotRepository,
	workshopRepo repository.WorkshopRepository,
	userRepo repository.UserRepository,
) OrderService {
	return &orderService{orderRepo, slotRepo, workshopRepo, userRepo}
}

func (s *orderService) Create(customerID uuid.UUID, req *domain.CreateOrderRequest) (*domain.Order, error) {
	workshopID, err := uuid.Parse(req.WorkshopID)
	if err != nil {
		return nil, errors.New("invalid workshop_id")
	}
	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		return nil, errors.New("invalid slot_id")
	}

	workshop, err := s.workshopRepo.FindByID(workshopID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("workshop not found")
	}
	if err != nil {
		return nil, err
	}

	slot, err := s.slotRepo.FindByID(slotID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("slot not found")
	}
	if err != nil {
		return nil, err
	}
	if slot.WorkshopID != workshopID {
		return nil, errors.New("slot does not belong to this workshop")
	}
	if !slot.IsAvailable() {
		return nil, errors.New("slot is full or already passed")
	}

	order := &domain.Order{
		CustomerID:   customerID,
		WorkshopID:   workshopID,
		SlotID:       slotID,
		VehicleType:  req.VehicleType,
		VehiclePlate: req.VehiclePlate,
		Notes:        req.Notes,
		Status:       domain.BookingStatusPending,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	if err := s.slotRepo.IncrementBooked(slotID); err != nil {
		return nil, err
	}

	// Notif WA ke operator workshop
	go func() {
		customer, err := s.userRepo.FindByID(customerID)
		if err != nil {
			return
		}
		operator, err := s.userRepo.FindByID(workshop.OwnerID)
		if err != nil {
			return
		}
		msg := fonnte.MsgNewOrder(
			customer.Name,
			order.VehicleType,
			order.VehiclePlate,
			workshop.Name,
			slot.Date,
			order.Notes,
		)
		fonnte.SendAsync(operator.Phone, msg)
	}()

	return order, nil
}

func (s *orderService) GetByID(id uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role) (*domain.Order, error) {
	order, err := s.orderRepo.FindByIDWithRelations(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	if requesterRole == domain.RoleCustomer && order.CustomerID != requesterID {
		return nil, errors.New("forbidden")
	}

	if requesterRole == domain.RoleOperator {
		workshop, err := s.workshopRepo.FindByID(order.WorkshopID)
		if err != nil || workshop.OwnerID != requesterID {
			return nil, errors.New("forbidden")
		}
	}

	return order, nil
}

func (s *orderService) GetMyOrders(customerID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.orderRepo.FindByCustomerID(customerID, page, limit)
}

func (s *orderService) GetWorkshopOrders(workshopID uuid.UUID, ownerID uuid.UUID, page, limit int) ([]domain.Order, int64, error) {
	workshop, err := s.workshopRepo.FindByID(workshopID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, errors.New("workshop not found")
	}
	if err != nil {
		return nil, 0, err
	}
	if workshop.OwnerID != ownerID {
		return nil, 0, errors.New("forbidden: you don't own this workshop")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.orderRepo.FindByWorkshopID(workshopID, page, limit)
}

func (s *orderService) UpdateStatus(id uuid.UUID, ownerID uuid.UUID, req *domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := s.orderRepo.FindByIDWithRelations(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	workshop, err := s.workshopRepo.FindByID(order.WorkshopID)
	if err != nil || workshop.OwnerID != ownerID {
		return nil, errors.New("forbidden: you don't own this workshop")
	}

	if err := validateStatusTransition(order.Status, req.Status); err != nil {
		return nil, err
	}

	if err := s.orderRepo.UpdateStatus(id, req.Status); err != nil {
		return nil, err
	}

	if req.Status == domain.BookingStatusCancelled {
		_ = s.slotRepo.DecrementBooked(order.SlotID)
	}

	// Notif WA ke customer
	go func() {
		customer, err := s.userRepo.FindByID(order.CustomerID)
		if err != nil {
			return
		}
		slot, err := s.slotRepo.FindByID(order.SlotID)
		if err != nil {
			return
		}
		msg := fonnte.MsgStatusUpdate(
			customer.Name,
			workshop.Name,
			order.VehiclePlate,
			string(req.Status),
			slot.Date,
		)
		fonnte.SendAsync(customer.Phone, msg)
	}()

	order.Status = req.Status
	return order, nil
}

func (s *orderService) Cancel(id uuid.UUID, customerID uuid.UUID) error {
	order, err := s.orderRepo.FindByIDWithRelations(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("order not found")
	}
	if err != nil {
		return err
	}

	if order.CustomerID != customerID {
		return errors.New("forbidden")
	}
	if order.Status != domain.BookingStatusPending {
		return errors.New("only pending orders can be cancelled")
	}

	if err := s.orderRepo.UpdateStatus(id, domain.BookingStatusCancelled); err != nil {
		return err
	}

	_ = s.slotRepo.DecrementBooked(order.SlotID)

	// Notif WA ke operator
	go func() {
		customer, err := s.userRepo.FindByID(customerID)
		if err != nil {
			return
		}
		workshop, err := s.workshopRepo.FindByID(order.WorkshopID)
		if err != nil {
			return
		}
		operator, err := s.userRepo.FindByID(workshop.OwnerID)
		if err != nil {
			return
		}
		slot, err := s.slotRepo.FindByID(order.SlotID)
		if err != nil {
			return
		}
		msg := fonnte.MsgOrderCancelled(
			operator.Name,
			customer.Name,
			order.VehicleType,
			order.VehiclePlate,
			workshop.Name,
			slot.Date,
		)
		fonnte.SendAsync(operator.Phone, msg)
	}()

	return nil
}

func validateStatusTransition(current, next domain.BookingStatus) error {
	allowed := map[domain.BookingStatus][]domain.BookingStatus{
		domain.BookingStatusPending:   {domain.BookingStatusConfirmed, domain.BookingStatusCancelled},
		domain.BookingStatusConfirmed: {domain.BookingStatusDone, domain.BookingStatusCancelled},
		domain.BookingStatusDone:      {},
		domain.BookingStatusCancelled: {},
	}

	for _, s := range allowed[current] {
		if s == next {
			return nil
		}
	}
	return errors.New("invalid status transition: " + string(current) + " → " + string(next))
}
