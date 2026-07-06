package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(vehicle *domain.Vehicle) error
	FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error)
	FindByCustomerID(customerID uuid.UUID) ([]domain.Vehicle, error)
	FindByID(id uuid.UUID) (*domain.Vehicle, error)
	FindByPlateNumber(workshopID uuid.UUID, plateNumber string) (*domain.Vehicle, error)
	FindByCustomerAndPlate(workshopID, customerID uuid.UUID, plateNumber string) (*domain.Vehicle, error)
	Update(vehicle *domain.Vehicle) error
	Delete(id uuid.UUID) error
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db}
}

func (r *vehicleRepository) Create(vehicle *domain.Vehicle) error {
	return r.db.Create(vehicle).Error
}

func (r *vehicleRepository) FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Vehicle, int64, error) {
	var vehicles []domain.Vehicle
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Vehicle{}).Where("workshop_id = ?", workshopID)

	if search != "" {
		query = query.Where("plate_number ILIKE ? OR brand ILIKE ? OR model ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Customer").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&vehicles).Error

	return vehicles, total, err
}

func (r *vehicleRepository) FindByCustomerID(customerID uuid.UUID) ([]domain.Vehicle, error) {
	var vehicles []domain.Vehicle
	err := r.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindByID(id uuid.UUID) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.Preload("Customer").First(&vehicle, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// FindByPlateNumber dipakai buat cari vehicle existing (dalam satu workshop) berdasarkan
// plat nomor doang — HATI-HATI: ini gak peduli siapa pemiliknya, jadi kalau plat yang sama
// kebetulan udah ada punya customer lain, method ini bakal ketemu punya customer itu juga.
// Buat proses convert booking → servis, pakai FindByCustomerAndPlate() di bawah, BUKAN ini.
func (r *vehicleRepository) FindByPlateNumber(workshopID uuid.UUID, plateNumber string) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.Where("workshop_id = ? AND plate_number = ?", workshopID, plateNumber).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// FindByCustomerAndPlate cari vehicle berdasarkan plat nomor YANG JUGA harus dimiliki
// customer yang sama. Ini sengaja di-scope ke customer_id juga (bukan cuma plate_number)
// supaya kalau ada 2 customer berbeda yang kebetulan input plat yang sama (typo, data
// testing, dll), booking/servis salah satu customer gak "nyasar" ke riwayat kendaraan
// customer lain yang gak related — vehicle baru bakal dibuatkan khusus buat customer ini.
func (r *vehicleRepository) FindByCustomerAndPlate(workshopID, customerID uuid.UUID, plateNumber string) (*domain.Vehicle, error) {
	var vehicle domain.Vehicle
	err := r.db.Where("workshop_id = ? AND customer_id = ? AND plate_number = ?", workshopID, customerID, plateNumber).First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) Update(vehicle *domain.Vehicle) error {
	return r.db.Save(vehicle).Error
}

func (r *vehicleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Vehicle{}, "id = ?", id).Error
}