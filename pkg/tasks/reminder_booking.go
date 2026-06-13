package tasks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Payload untuk task reminder booking H-1
type ReminderBookingPayload struct {
	OrderID      uuid.UUID `json:"order_id"`
	CustomerName string    `json:"customer_name"`
	CustomerPhone string   `json:"customer_phone"`
	WorkshopName string    `json:"workshop_name"`
	VehiclePlate string    `json:"vehicle_plate"`
	SlotDate     time.Time `json:"slot_date"`
}

// NewReminderBookingTask buat task baru, dijadwalkan H-1 sebelum slot
func NewReminderBookingTask(payload ReminderBookingPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeReminderBooking, data), nil
}
