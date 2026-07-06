package fonnte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/naufalnak/bengkelhub-backend/config"
)

const fonnteURL = "https://api.fonnte.com/send"

type sendRequest struct {
	Target  string `json:"target"`
	Message string `json:"message"`
}

type sendResponse struct {
	Status  bool   `json:"status"`
	Reason  string `json:"reason"`  // Fonnte balikin alasan gagal di field "reason", BUKAN "message"
	Message string `json:"message"` // beberapa endpoint/versi Fonnte lain pakai "message" — disimpan juga buat jaga-jaga
}

// Send kirim pesan WA ke nomor target (format: 628xxx)
func Send(phone, message string) error {
	if config.Cfg.FonnteToken == "" {
		log.Println("[Fonnte] Token not set, skipping WA notification")
		return nil
	}
	if phone == "" {
		log.Println("[Fonnte] Empty phone number, skipping")
		return nil
	}

	payload, _ := json.Marshal(sendRequest{
		Target:  normalizePhone(phone),
		Message: message,
	})

	req, err := http.NewRequest(http.MethodPost, fonnteURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", config.Cfg.FonnteToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result sendResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse Fonnte response: %w", err)
	}

	if !result.Status {
		reason := result.Reason
		if reason == "" {
			reason = result.Message
		}
		if reason == "" {
			// Jaga-jaga kalau Fonnte suatu saat ganti lagi nama field-nya —
			// biar errornya tetap informatif (nunjukin raw response), bukan kosong.
			reason = string(body)
		}
		return fmt.Errorf("Fonnte error: %s", reason)
	}

	log.Printf("[Fonnte] Message sent to %s", phone)
	return nil
}

// SendAsync kirim notif secara async (non-blocking) — dipakai di service
func SendAsync(phone, message string) {
	go func() {
		if err := Send(phone, message); err != nil {
			log.Printf("[Fonnte] Failed to send to %s: %v", phone, err)
		}
	}()
}

// normalizePhone: +6281234567890 → 6281234567890
func normalizePhone(phone string) string {
	if len(phone) > 0 && phone[0] == '+' {
		return phone[1:]
	}
	return phone
}