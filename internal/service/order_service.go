package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/pkg/fonnte"
	"github.com/naufalnak/bengkelhub-backend/pkg/tasks"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(customerID uuid.UUID, req *domain.CreateOrderRequest) (*domain.Order, error)
	GetByID(id uuid.UUID, requesterID uuid.UUID, requesterRole domain.Role) (*domain.Order, error)
	GetMyOrders(customerID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	GetWorkshopOrders(workshopID uuid.UUID, ownerID uuid.UUID, page, limit int) ([]domain.Order, int64, error)
	UpdateStatus(id uuid.UUID, ownerID uuid.UUID, req *domain.UpdateOrderStatusRequest) (*domain.Order, error)
	Cancel(id uuid.UUID, customerID uuid.UUID) error
	ConvertToService(orderID, ownerID uuid.UUID) (*domain.Service, error)
}

type orderService struct {
	orderRepo    repository.OrderRepository
	slotRepo     repository.SlotRepository
	workshopRepo repository.WorkshopRepository
	userRepo     repository.UserRepository
	customerRepo repository.CustomerRepository
	vehicleRepo  repository.VehicleRepository
	serviceRepo  repository.ServiceRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	slotRepo repository.SlotRepository,
	workshopRepo repository.WorkshopRepository,
	userRepo repository.UserRepository,
	customerRepo repository.CustomerRepository,
	vehicleRepo repository.VehicleRepository,
	serviceRepo repository.ServiceRepository,
) OrderService {
	return &orderService{orderRepo, slotRepo, workshopRepo, userRepo, customerRepo, vehicleRepo, serviceRepo}
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

	// Auto-sync ke Customer & Vehicle internal begitu booking masuk — biar
	// operator langsung lihat datanya di menu Pelanggan & Kendaraan tanpa
	// perlu klik "Proses jadi Servis" dulu. Sengaja gak fatal kalau gagal
	// (mis. customer belum punya nomor HP) — booking-nya tetap harus sukses,
	// nanti operator masih bisa lengkapi manual pas convert ke Service.
	if customerUser, err := s.userRepo.FindByID(customerID); err != nil {
		log.Printf("[Order] Failed to load customer %s for auto-sync: %v", customerID, err)
	} else if _, _, err := s.findOrCreateCustomerAndVehicle(
		workshopID,
		customerUser.Name, customerUser.Phone, customerUser.Email,
		req.VehiclePlate, req.VehicleType,
	); err != nil {
		log.Printf("[Order] Auto-sync customer/vehicle for order %s failed: %v", order.ID, err)
	}

	// Notif + enqueue reminder (async)
	go func() {
		customer, err := s.userRepo.FindByID(customerID)
		if err != nil {
			return
		}
		operator, err := s.userRepo.FindByID(workshop.OwnerID)
		if err != nil {
			return
		}

		// WA ke operator — booking baru
		msg := fonnte.MsgNewOrder(
			customer.Name, order.VehicleType, order.VehiclePlate,
			workshop.Name, slot.Date, order.Notes,
		)
		fonnte.SendAsync(operator.Phone, msg)

		// Enqueue reminder H-1 ke customer
		if err := tasks.EnqueueReminderBooking(tasks.ReminderBookingPayload{
			OrderID:       order.ID,
			CustomerName:  customer.Name,
			CustomerPhone: customer.Phone,
			WorkshopName:  workshop.Name,
			VehiclePlate:  order.VehiclePlate,
			SlotDate:      slot.Date,
		}); err != nil {
			log.Printf("[Order] Failed to enqueue reminder: %v", err)
		}
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
			customer.Name, workshop.Name, order.VehiclePlate,
			string(req.Status), slot.Date,
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
			operator.Name, customer.Name, order.VehicleType,
			order.VehiclePlate, workshop.Name, slot.Date,
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

// findOrCreateCustomerAndVehicle nyari (atau bikin baru) Customer & Vehicle
// internal berdasarkan nomor HP & plat nomor. Dipakai di dua tempat:
//  1. Order.Create() — begitu booking masuk, biar langsung kelihatan di menu
//     Pelanggan & Kendaraan tanpa nunggu operator klik apa-apa.
//  2. Order.ConvertToService() — biar idempotent kalau dipanggil lagi (Customer/
//     Vehicle-nya kemungkinan besar sudah ada dari langkah 1 di atas).
func (s *orderService) findOrCreateCustomerAndVehicle(
	workshopID uuid.UUID,
	name, phone, email, plateNumber, vehicleType string,
) (*domain.Customer, *domain.Vehicle, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, nil, errors.New("customer phone not available")
	}

	// 1. Cari atau buat Customer (internal), berdasarkan nomor HP
	customer, err := s.customerRepo.FindByPhone(workshopID, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		customer = &domain.Customer{
			WorkshopID: workshopID,
			Name:       name,
			Phone:      phone,
			Email:      email,
		}
		if err := s.customerRepo.Create(customer); err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	// 2. Cari atau buat Vehicle — di-scope ke plat nomor YANG JUGA milik customer ini.
	// Sengaja BUKAN pakai FindByPlateNumber (yang cuma cek plat doang di seluruh
	// workshop) — soalnya kalau ada customer lain yang kebetulan punya kendaraan
	// dengan plat yang sama, kita gak mau riwayat servis/booking customer ini
	// "nyasar" nempel ke kendaraan customer lain itu. Kalau plat yang sama ternyata
	// punya customer berbeda, di sini vehicle baru tetap dibuatkan khusus buat
	// customer ini (operator bisa gabungin manual nanti kalau itu memang kendaraan
	// yang sama/pindah tangan).
	vehicle, err := s.vehicleRepo.FindByCustomerAndPlate(workshopID, customer.ID, plateNumber)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		vehicle = &domain.Vehicle{
			WorkshopID:  workshopID,
			CustomerID:  customer.ID,
			PlateNumber: plateNumber,
			// vehicleType diisi bebas oleh customer saat booking (mis. "FreeGo",
			// "Honda Beat 2022") — belum tentu pas dipisah jadi Brand/Model, jadi
			// sementara ditaruh semua di Model. Operator bisa rapikan lagi manual nanti.
			Model: vehicleType,
		}
		if err := s.vehicleRepo.Create(vehicle); err != nil {
			return nil, nil, err
		}
	} else if err != nil {
		return nil, nil, err
	}

	return customer, vehicle, nil
}

// ConvertToService mengubah sebuah Order (booking) jadi Service internal.
// Customer & Vehicle-nya sendiri normalnya SUDAH otomatis kebuat sejak booking
// pertama kali masuk (lihat Create()) — method ini tinggal pastikan lagi
// (idempotent) lalu bikinkan Service-nya.
//
// Order yang sudah pernah dikonversi (Order.ServiceID != nil) gak bisa dikonversi lagi —
// biar gak numpuk Service duplikat kalau operator klik tombolnya berkali-kali.
func (s *orderService) ConvertToService(orderID, ownerID uuid.UUID) (*domain.Service, error) {
	order, err := s.orderRepo.FindByIDWithRelations(orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}

	if _, err := verifyWorkshopOwner(s.workshopRepo, order.WorkshopID, ownerID); err != nil {
		return nil, err
	}

	if order.ServiceID != nil {
		return nil, errors.New("order already converted to service")
	}
	if order.Status == domain.BookingStatusCancelled {
		return nil, errors.New("cannot convert a cancelled order")
	}

	_, vehicle, err := s.findOrCreateCustomerAndVehicle(
		order.WorkshopID,
		order.Customer.Name, order.Customer.Phone, order.Customer.Email,
		order.VehiclePlate, order.VehicleType,
	)
	if err != nil {
		return nil, err
	}

	// Buat Service baru, linked ke Vehicle di atas
	complaint := strings.TrimSpace(order.Notes)
	if complaint == "" {
		complaint = "Booking dari aplikasi - " + order.VehicleType
	}

	svc := &domain.Service{
		WorkshopID: order.WorkshopID,
		VehicleID:  vehicle.ID,
		ServiceNo:  generateServiceNo(),
		Complaint:  complaint,
		Status:     domain.ServiceStatusPending,
		StartDate:  time.Now(),
	}
	if err := s.serviceRepo.Create(svc); err != nil {
		return nil, err
	}

	// Tandai order ini sudah dikonversi, biar gak dobel
	if err := s.orderRepo.UpdateServiceID(order.ID, svc.ID); err != nil {
		log.Printf("[Order] Failed to link order %s to service %s: %v", order.ID, svc.ID, err)
	}

	return svc, nil
}