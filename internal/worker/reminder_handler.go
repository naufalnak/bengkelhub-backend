package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/naufalnak/bengkelhub-backend/pkg/fonnte"
	"github.com/naufalnak/bengkelhub-backend/pkg/tasks"
)

// HandleReminderBooking proses task reminder H-1
func HandleReminderBooking(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ReminderBookingPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[Worker] Sending reminder to %s for order %s", payload.CustomerPhone, payload.OrderID)

	msg := buildReminderMessage(payload)
	if err := fonnte.Send(payload.CustomerPhone, msg); err != nil {
		// Return error supaya Asynq retry otomatis
		return fmt.Errorf("failed to send WA reminder: %w", err)
	}

	log.Printf("[Worker] Reminder sent successfully for order %s", payload.OrderID)
	return nil
}

func buildReminderMessage(p tasks.ReminderBookingPayload) string {
	wib, _ := time.LoadLocation("Asia/Jakarta")
	jadwal := p.SlotDate.In(wib).Format("02 Jan 2006, 15:04 WIB")

	return fmt.Sprintf(
		"⏰ *Reminder Servis Besok!*\n\n"+
			"Halo %s,\n\n"+
			"Jangan lupa, kamu punya jadwal servis *besok*:\n\n"+
			"🏪 Bengkel: *%s*\n"+
			"🚗 Kendaraan: *%s*\n"+
			"📅 Jadwal: *%s*\n\n"+
			"Pastikan datang tepat waktu ya! 🙏",
		p.CustomerName,
		p.WorkshopName,
		p.VehiclePlate,
		jadwal,
	)
}
