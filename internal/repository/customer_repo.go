package repository

import (
	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"gorm.io/gorm"
)

type CustomerRepository interface {
	Create(customer *domain.Customer) error
	FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Customer, int64, error)
	FindByID(id uuid.UUID) (*domain.Customer, error)
	FindByPhone(workshopID uuid.UUID, phone string) (*domain.Customer, error)
	Update(customer *domain.Customer) error
	Delete(id uuid.UUID) error
}

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{db}
}

func (r *customerRepository) Create(customer *domain.Customer) error {
	return r.db.Create(customer).Error
}

func (r *customerRepository) FindByWorkshopID(workshopID uuid.UUID, search string, page, limit int) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&domain.Customer{}).Where("workshop_id = ?", workshopID)

	if search != "" {
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&customers).Error

	return customers, total, err
}

func (r *customerRepository) FindByID(id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.First(&customer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindByPhone dipakai buat cari customer existing (dalam satu workshop) berdasarkan
// nomor HP — dipakai saat konversi Order jadi Service biar gak bikin duplikat.
func (r *customerRepository) FindByPhone(workshopID uuid.UUID, phone string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.Where("workshop_id = ? AND phone = ?", workshopID, phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(customer *domain.Customer) error {
	return r.db.Save(customer).Error
}

func (r *customerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Customer{}, "id = ?", id).Error
}