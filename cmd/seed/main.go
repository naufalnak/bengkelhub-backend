package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

func hashPassword(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt error: %v", err)
	}
	return string(b)
}

func ptr[T any](v T) *T { return &v }

func serviceNo() string {
	return fmt.Sprintf("SRV-%s-%04d", time.Now().Format("20060102"), rand.Intn(9999)+1)
}

func invoiceNo() string {
	return fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), rand.Intn(9999)+1)
}

// ── Main ─────────────────────────────────────────────────────────────────────

func main() {
	config.Load()
	config.ConnectDB()
	db := config.DB

	// Kalau ada flag --fresh, hapus semua data dulu sebelum seed
	fresh := len(os.Args) > 1 && os.Args[1] == "--fresh"
	if fresh {
		log.Println("⚠️   Flag --fresh terdeteksi — menghapus semua data terlebih dahulu...")
		wipeAll(db)
		log.Println("✅  Database bersih, mulai seed ulang...")
	}

	log.Println("🌱  Starting seed...")

	// ── 1. Operators ─────────────────────────────────────────────────────────
	log.Println("→ Creating operators...")

	op1 := &domain.User{
		Name:          "Budi Santoso",
		Email:         "budi@bengkelhub.test",
		Password:      hashPassword("password123"),
		Phone:         "+6281234567890",
		Role:          domain.RoleOperator,
		EmailVerified: true,
	}
	op2 := &domain.User{
		Name:          "Siti Operator",
		Email:         "siti@bengkelhub.test",
		Password:      hashPassword("password123"),
		Phone:         "+6289876543210",
		Role:          domain.RoleOperator,
		EmailVerified: true,
	}

	upsertUser(db, op1)
	upsertUser(db, op2)

	// ── 2. Customers (akun login) ─────────────────────────────────────────────
	log.Println("→ Creating customer accounts...")

	cust1 := &domain.User{
		Name:          "Andi Customer",
		Email:         "andi@bengkelhub.test",
		Password:      hashPassword("password123"),
		Phone:         "+6281111111111",
		Role:          domain.RoleCustomer,
		EmailVerified: true,
	}

	upsertUser(db, cust1)

	// ── 3. Workshops ──────────────────────────────────────────────────────────
	log.Println("→ Creating workshops...")

	ws1 := &domain.Workshop{
		OwnerID:     op1.ID,
		Name:        "Bengkel Jaya Motor",
		Description: "Bengkel umum motor dan mobil terpercaya di Bekasi",
		Address:     "Jl. Raya Bekasi No. 10, Bekasi Utara",
		Phone:       "+622112345678",
		IsActive:    true,
	}
	ws2 := &domain.Workshop{
		OwnerID:     op2.ID,
		Name:        "Auto Care Cikarang",
		Description: "Spesialis servis AC, tune up, dan perawatan berkala",
		Address:     "Jl. Industri Raya No. 45, Cikarang Selatan",
		Phone:       "+622187654321",
		IsActive:    true,
	}

	if err := db.Where("name = ?", ws1.Name).FirstOrCreate(ws1).Error; err != nil {
		log.Fatalf("ws1 error: %v", err)
	}
	if err := db.Where("name = ?", ws2.Name).FirstOrCreate(ws2).Error; err != nil {
		log.Fatalf("ws2 error: %v", err)
	}

	// ── 4. Slots ──────────────────────────────────────────────────────────────
	log.Println("→ Creating slots...")

	now := time.Now()
	slotDates := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, time.Local),
		time.Date(now.Year(), now.Month(), now.Day()+1, 11, 0, 0, 0, time.Local),
		time.Date(now.Year(), now.Month(), now.Day()+2, 9, 0, 0, 0, time.Local),
		time.Date(now.Year(), now.Month(), now.Day()+3, 10, 0, 0, 0, time.Local),
	}
	for _, d := range slotDates {
		slot := &domain.Slot{WorkshopID: ws1.ID, Date: d, MaxBooking: 5, Booked: 0}
		db.Where("workshop_id = ? AND date = ?", slot.WorkshopID, slot.Date).FirstOrCreate(slot)
	}
	// Slot ws2
	slot2 := &domain.Slot{
		WorkshopID: ws2.ID,
		Date:       time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, time.Local),
		MaxBooking: 3,
		Booked:     0,
	}
	db.Where("workshop_id = ? AND date = ?", slot2.WorkshopID, slot2.Date).FirstOrCreate(slot2)

	// ── 5. Walk-in Customers (data internal bengkel, bukan akun login) ────────
	log.Println("→ Creating walk-in customers...")

	wCustomers := []domain.Customer{
		{WorkshopID: ws1.ID, Name: "Hendra Gunawan", Phone: "+6281234000001", Email: "hendra@email.com", Address: "Jl. Kemang Raya No. 5"},
		{WorkshopID: ws1.ID, Name: "Dewi Lestari", Phone: "+6281234000002", Email: "dewi@email.com", Address: "Jl. Sudirman No. 12"},
		{WorkshopID: ws1.ID, Name: "Rudi Hartono", Phone: "+6281234000003", Email: "", Address: "Jl. Gatot Subroto No. 8"},
		{WorkshopID: ws2.ID, Name: "Yuli Pratiwi", Phone: "+6281234000004", Email: "yuli@email.com", Address: "Jl. Industri No. 20"},
		{WorkshopID: ws2.ID, Name: "Doni Setiawan", Phone: "+6281234000005", Email: "", Address: "Jl. Pahlawan No. 3"},
	}
	for i := range wCustomers {
		db.Where("workshop_id = ? AND phone = ?", wCustomers[i].WorkshopID, wCustomers[i].Phone).
			FirstOrCreate(&wCustomers[i])
	}

	// ── 6. Vehicles ───────────────────────────────────────────────────────────
	log.Println("→ Creating vehicles...")

	vehicles := []domain.Vehicle{
		{WorkshopID: ws1.ID, CustomerID: wCustomers[0].ID, PlateNumber: "B 1234 ABC", Brand: "Honda", Model: "Vario 150", Year: 2022, Color: "Hitam", EngineCC: 150},
		{WorkshopID: ws1.ID, CustomerID: wCustomers[0].ID, PlateNumber: "B 5678 DEF", Brand: "Yamaha", Model: "NMAX 155", Year: 2021, Color: "Biru", EngineCC: 155},
		{WorkshopID: ws1.ID, CustomerID: wCustomers[1].ID, PlateNumber: "B 9012 GHI", Brand: "Toyota", Model: "Avanza", Year: 2019, Color: "Putih", EngineCC: 1300},
		{WorkshopID: ws1.ID, CustomerID: wCustomers[2].ID, PlateNumber: "B 3456 JKL", Brand: "Honda", Model: "Beat Street", Year: 2023, Color: "Merah", EngineCC: 110},
		{WorkshopID: ws2.ID, CustomerID: wCustomers[3].ID, PlateNumber: "T 1111 AAA", Brand: "Suzuki", Model: "Ertiga", Year: 2020, Color: "Silver", EngineCC: 1500},
		{WorkshopID: ws2.ID, CustomerID: wCustomers[4].ID, PlateNumber: "T 2222 BBB", Brand: "Daihatsu", Model: "Xenia", Year: 2018, Color: "Abu-abu", EngineCC: 1300},
	}
	for i := range vehicles {
		db.Where("workshop_id = ? AND plate_number = ?", vehicles[i].WorkshopID, vehicles[i].PlateNumber).
			FirstOrCreate(&vehicles[i])
	}

	// ── 7. Services ───────────────────────────────────────────────────────────
	log.Println("→ Creating services...")

	// Service 1: sudah selesai, lengkap dengan invoice & payment
	svc1 := &domain.Service{
		WorkshopID: ws1.ID,
		VehicleID:  vehicles[0].ID,
		ServiceNo:  "SRV-DEMO-0001",
		Complaint:  "Suara mesin kasar dan oli perlu diganti",
		Diagnosis:  "Oli mesin habis, kampas rem depan tipis",
		Notes:      "Customer minta cek juga ban depan",
		Status:     domain.ServiceStatusDone,
		StartDate:  time.Now().Add(-48 * time.Hour),
		EndDate:    ptr(time.Now().Add(-24 * time.Hour)),
	}
	db.Where("service_no = ?", svc1.ServiceNo).FirstOrCreate(svc1)

	// Service 1 items
	items1 := []domain.ServiceItem{
		{ServiceID: svc1.ID, Name: "Oli Mesin Shell Helix", Description: "1L SAE 10W-40", Qty: 1, UnitPrice: 75000, Total: 75000},
		{ServiceID: svc1.ID, Name: "Kampas Rem Depan", Description: "Original Honda", Qty: 1, UnitPrice: 85000, Total: 85000},
		{ServiceID: svc1.ID, Name: "Jasa Servis", Description: "Ganti oli + cek rem", Qty: 1, UnitPrice: 50000, Total: 50000},
	}
	for i := range items1 {
		db.Where("service_id = ? AND name = ?", items1[i].ServiceID, items1[i].Name).FirstOrCreate(&items1[i])
	}

	// Invoice untuk svc1 (sudah lunas)
	inv1 := &domain.Invoice{
		WorkshopID: ws1.ID,
		ServiceID:  svc1.ID,
		InvoiceNo:  "INV-DEMO-0001",
		Subtotal:   210000,
		Tax:        0,
		Discount:   0,
		Total:      210000,
		Status:     domain.InvoiceStatusPaid,
		DueDate:    ptr(time.Now().Add(7 * 24 * time.Hour)),
	}
	db.Where("invoice_no = ?", inv1.InvoiceNo).FirstOrCreate(inv1)

	pay1 := &domain.Payment{
		WorkshopID: ws1.ID,
		InvoiceID:  inv1.ID,
		Amount:     210000,
		Method:     domain.PaymentMethodCash,
		Notes:      "Bayar tunai lunas",
		PaidAt:     time.Now().Add(-20 * time.Hour),
	}
	db.Where("invoice_id = ? AND amount = ?", pay1.InvoiceID, pay1.Amount).FirstOrCreate(pay1)

	// Service 2: sedang dikerjakan, ada items, belum invoice
	svc2 := &domain.Service{
		WorkshopID: ws1.ID,
		VehicleID:  vehicles[2].ID,
		ServiceNo:  "SRV-DEMO-0002",
		Complaint:  "AC tidak dingin dan mesin overheat",
		Diagnosis:  "Freon AC habis, kipas radiator mati",
		Status:     domain.ServiceStatusInProgress,
		StartDate:  time.Now().Add(-2 * time.Hour),
	}
	db.Where("service_no = ?", svc2.ServiceNo).FirstOrCreate(svc2)

	items2 := []domain.ServiceItem{
		{ServiceID: svc2.ID, Name: "Freon AC R32", Qty: 1, UnitPrice: 150000, Total: 150000},
		{ServiceID: svc2.ID, Name: "Jasa Isi Freon", Qty: 1, UnitPrice: 75000, Total: 75000},
	}
	for i := range items2 {
		db.Where("service_id = ? AND name = ?", items2[i].ServiceID, items2[i].Name).FirstOrCreate(&items2[i])
	}

	// Service 3: baru masuk, pending
	svc3 := &domain.Service{
		WorkshopID: ws1.ID,
		VehicleID:  vehicles[3].ID,
		ServiceNo:  "SRV-DEMO-0003",
		Complaint:  "Ganti ban dan tune up rutin",
		Status:     domain.ServiceStatusPending,
		StartDate:  time.Now(),
	}
	db.Where("service_no = ?", svc3.ServiceNo).FirstOrCreate(svc3)

	// Service 4: di workshop 2, sudah selesai, partial payment
	svc4 := &domain.Service{
		WorkshopID: ws2.ID,
		VehicleID:  vehicles[4].ID,
		ServiceNo:  "SRV-DEMO-0004",
		Complaint:  "Tune up dan ganti filter udara",
		Diagnosis:  "Filter udara kotor, busi perlu diganti",
		Status:     domain.ServiceStatusDone,
		StartDate:  time.Now().Add(-72 * time.Hour),
		EndDate:    ptr(time.Now().Add(-48 * time.Hour)),
	}
	db.Where("service_no = ?", svc4.ServiceNo).FirstOrCreate(svc4)

	items4 := []domain.ServiceItem{
		{ServiceID: svc4.ID, Name: "Filter Udara", Qty: 1, UnitPrice: 65000, Total: 65000},
		{ServiceID: svc4.ID, Name: "Busi NGK", Qty: 4, UnitPrice: 35000, Total: 140000},
		{ServiceID: svc4.ID, Name: "Jasa Tune Up", Qty: 1, UnitPrice: 100000, Total: 100000},
	}
	for i := range items4 {
		db.Where("service_id = ? AND name = ?", items4[i].ServiceID, items4[i].Name).FirstOrCreate(&items4[i])
	}

	// Invoice ws2 (sebagian dibayar)
	inv2 := &domain.Invoice{
		WorkshopID: ws2.ID,
		ServiceID:  svc4.ID,
		InvoiceNo:  "INV-DEMO-0002",
		Subtotal:   305000,
		Tax:        0,
		Discount:   5000,
		Total:      300000,
		Status:     domain.InvoiceStatusPartial,
	}
	db.Where("invoice_no = ?", inv2.InvoiceNo).FirstOrCreate(inv2)

	pay2 := &domain.Payment{
		WorkshopID:  ws2.ID,
		InvoiceID:   inv2.ID,
		Amount:      150000,
		Method:      domain.PaymentMethodTransfer,
		ReferenceNo: "TRF20260101001",
		Notes:       "DP via transfer",
		PaidAt:      time.Now().Add(-45 * time.Hour),
	}
	db.Where("invoice_id = ? AND reference_no = ?", pay2.InvoiceID, pay2.ReferenceNo).FirstOrCreate(pay2)

	// ── 8. Booking orders (publik) ────────────────────────────────────────────
	log.Println("→ Creating booking orders...")

	order1 := &domain.Order{
		CustomerID:   cust1.ID,
		WorkshopID:   ws1.ID,
		SlotID:       slot2.ID,
		Status:       "confirmed",
		Notes:        "Ganti oli dan cek rem belakang",
		VehicleType:  "Honda Beat 2022",
		VehiclePlate: "B 9999 ZZZ",
	}
	db.Where("customer_id = ? AND workshop_id = ? AND vehicle_plate = ?",
		order1.CustomerID, order1.WorkshopID, order1.VehiclePlate).FirstOrCreate(order1)

	// ── Done ──────────────────────────────────────────────────────────────────

	log.Println("")
	log.Println("✅  Seeding selesai! Berikut akun yang dibuat:")
	log.Println("")
	log.Println("┌─────────────────────────────────────────────────────┐")
	log.Println("│  OPERATOR                                           │")
	log.Printf("│  Email : %-40s │", op1.Email)
	log.Printf("│  Email : %-40s │", op2.Email)
	log.Println("│  Password: password123                              │")
	log.Println("├─────────────────────────────────────────────────────┤")
	log.Println("│  CUSTOMER (akun login)                              │")
	log.Printf("│  Email : %-40s │", cust1.Email)
	log.Println("│  Password: password123                              │")
	log.Println("├─────────────────────────────────────────────────────┤")
	log.Println("│  DATA YANG DIBUAT                                   │")
	log.Println("│  - 2 workshop                                       │")
	log.Println("│  - 5 slot booking                                   │")
	log.Println("│  - 5 pelanggan walk-in                              │")
	log.Println("│  - 6 kendaraan                                      │")
	log.Println("│  - 4 servis (done/in_progress/pending)              │")
	log.Println("│  - 2 invoice (paid + partial)                       │")
	log.Println("│  - 2 payment                                        │")
	log.Println("│  - 1 booking online                                 │")
	log.Println("└─────────────────────────────────────────────────────┘")
}

func upsertUser(db *gorm.DB, u *domain.User) {
	existing := &domain.User{}
	if err := db.Where("email = ?", u.Email).First(existing).Error; err == nil {
		u.ID = existing.ID
		return
	}
	if err := db.Create(u).Error; err != nil {
		log.Fatalf("create user %s error: %v", u.Email, err)
	}
}

// wipeAll hapus semua data dari semua tabel — urutan penting karena foreign key
// (hapus child dulu sebelum parent)
func wipeAll(db *gorm.DB) {
	tables := []interface{}{
		&domain.Payment{},
		&domain.Invoice{},
		&domain.ServiceItem{},
		&domain.Service{},
		&domain.Order{},
		&domain.Slot{},
		&domain.Vehicle{},
		&domain.Customer{},
		&domain.Workshop{},
		&domain.User{},
	}

	for _, table := range tables {
		// TRUNCATE emulasi via DELETE WHERE 1=1 — aman buat semua DB
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			log.Printf("  ⚠️  Gagal hapus tabel %T: %v (mungkin sudah kosong)", table, err)
		} else {
			log.Printf("  🗑️  Tabel %T dihapus", table)
		}
	}
}