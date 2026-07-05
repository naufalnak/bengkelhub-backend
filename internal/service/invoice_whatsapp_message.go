package service

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/naufalnak/bengkelhub-backend/internal/domain"
)

var invoiceStatusLabelID = map[domain.InvoiceStatus]string{
	domain.InvoiceStatusUnpaid:  "Belum Dibayar",
	domain.InvoiceStatusPartial: "Dibayar Sebagian",
	domain.InvoiceStatusPaid:    "Lunas",
}

// buildInvoiceWhatsappMessage menyusun ringkasan invoice untuk dikirim via WA.
func buildInvoiceWhatsappMessage(invoice *domain.Invoice, workshopName, customerName string) string {
	var totalPaid float64
	for _, p := range invoice.Payments {
		totalPaid += p.Amount
	}
	remaining := invoice.Total - totalPaid
	if remaining < 0 {
		remaining = 0
	}

	statusLabel := invoiceStatusLabelID[invoice.Status]
	if statusLabel == "" {
		statusLabel = string(invoice.Status)
	}

	msg := fmt.Sprintf(
		"🧾 *Invoice dari %s*\n\n"+
			"Halo %s,\n\n"+
			"Berikut ringkasan invoice servis kendaraan kamu:\n\n"+
			"No. Invoice: *%s*\n"+
			"Total Tagihan: *%s*\n"+
			"Status: *%s*\n",
		workshopName, customerName, invoice.InvoiceNo, formatRupiah(invoice.Total), statusLabel,
	)

	if invoice.Status != domain.InvoiceStatusPaid && remaining > 0 {
		msg += fmt.Sprintf("Sisa Tagihan: *%s*\n", formatRupiah(remaining))
	}

	if invoice.PaymentURL != "" && invoice.Status != domain.InvoiceStatusPaid {
		msg += fmt.Sprintf("\n💳 Bayar online lewat link berikut:\n%s\n", invoice.PaymentURL)
	}

	msg += "\nTerima kasih sudah mempercayakan servis kendaraan Anda kepada kami 🙏"

	return msg
}

// formatRupiah format angka ke "Rp150.000" (pemisah ribuan pakai titik, tanpa desimal).
func formatRupiah(amount float64) string {
	n := int64(amount)
	neg := n < 0
	if neg {
		n = -n
	}

	s := strconv.FormatInt(n, 10)
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)

	result := "Rp" + strings.Join(parts, ".")
	if neg {
		result = "-" + result
	}
	return result
}