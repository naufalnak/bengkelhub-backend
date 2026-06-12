package fonnte

import (
	"fmt"
	"time"
)

// Pesan ke operator saat ada booking baru
func MsgNewOrder(customerName, vehicleType, vehiclePlate, workshopName string, slotDate time.Time, notes string) string {
	msg := fmt.Sprintf(
		"🔔 *Booking Baru - %s*\n\n"+
			"👤 Customer: %s\n"+
			"🚗 Kendaraan: %s (%s)\n"+
			"📅 Jadwal: %s\n",
		workshopName,
		customerName,
		vehicleType,
		vehiclePlate,
		slotDate.In(wibLocation()).Format("02 Jan 2006, 15:04 WIB"),
	)
	if notes != "" {
		msg += fmt.Sprintf("📝 Catatan: %s\n", notes)
	}
	msg += "\nSegera konfirmasi booking ini melalui dashboard."
	return msg
}

// Pesan ke customer saat status berubah
func MsgStatusUpdate(customerName, workshopName, vehiclePlate, status string, slotDate time.Time) string {
	emoji := statusEmoji(status)
	statusID := statusIndonesia(status)

	return fmt.Sprintf(
		"%s *Update Booking - %s*\n\n"+
			"Halo %s,\n\n"+
			"Status booking kendaraan *%s* kamu di *%s* telah diperbarui.\n\n"+
			"📅 Jadwal: %s\n"+
			"📌 Status: *%s*\n\n"+
			"%s",
		emoji,
		workshopName,
		customerName,
		vehiclePlate,
		workshopName,
		slotDate.In(wibLocation()).Format("02 Jan 2006, 15:04 WIB"),
		statusID,
		statusFooter(status),
	)
}

// Pesan ke operator saat customer cancel
func MsgOrderCancelled(operatorName, customerName, vehicleType, vehiclePlate, workshopName string, slotDate time.Time) string {
	return fmt.Sprintf(
		"❌ *Booking Dibatalkan - %s*\n\n"+
			"Customer *%s* telah membatalkan booking:\n\n"+
			"🚗 Kendaraan: %s (%s)\n"+
			"📅 Jadwal: %s\n\n"+
			"Slot telah otomatis tersedia kembali.",
		workshopName,
		customerName,
		vehicleType,
		vehiclePlate,
		slotDate.In(wibLocation()).Format("02 Jan 2006, 15:04 WIB"),
	)
}

func wibLocation() *time.Location {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return loc
}

func statusEmoji(status string) string {
	switch status {
	case "confirmed":
		return "✅"
	case "done":
		return "🎉"
	case "cancelled":
		return "❌"
	default:
		return "🔔"
	}
}

func statusIndonesia(status string) string {
	switch status {
	case "confirmed":
		return "Dikonfirmasi"
	case "done":
		return "Selesai"
	case "cancelled":
		return "Dibatalkan"
	default:
		return status
	}
}

func statusFooter(status string) string {
	switch status {
	case "confirmed":
		return "Silakan datang sesuai jadwal. Terima kasih! 🙏"
	case "done":
		return "Terima kasih telah menggunakan layanan kami! ⭐"
	case "cancelled":
		return "Booking kamu telah dibatalkan. Silakan booking ulang jika diperlukan."
	default:
		return ""
	}
}
