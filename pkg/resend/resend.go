package resend

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

const resendURL = "https://api.resend.com/emails"

type EmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type resendResponse struct {
	ID    string `json:"id"`
	Error string `json:"message"`
}

// Send kirim email transaksional lewat Resend API.
// Kalau RESEND_API_KEY kosong, skip dan log saja.
func Send(payload EmailPayload) error {
	if config.Cfg.ResendAPIKey == "" {
		log.Println("[Resend] API key not set, skipping email send")
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, resendURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+config.Cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errResp resendResponse
		_ = json.Unmarshal(raw, &errResp)
		return fmt.Errorf("resend error %d: %s", resp.StatusCode, errResp.Error)
	}

	var result resendResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("failed to parse Resend response: %w", err)
	}

	log.Printf("[Resend] Email sent id=%s to=%v", result.ID, payload.To)
	return nil
}

// SendVerificationEmail kirim email verifikasi ke user baru
func SendVerificationEmail(toEmail, toName, verifyLink string) error {
	return Send(EmailPayload{
		From:    "BengkelHub <noreply@bengkelhub.app>",
		To:      []string{toEmail},
		Subject: "Verifikasi Email BengkelHub",
		HTML:    buildVerificationHTML(toName, verifyLink),
	})
}

// SendAsync kirim email secara non-blocking — pakai goroutine biasa.
// Untuk production, ganti ini dengan Asyncq task biar retry-able.
func SendAsync(payload EmailPayload) {
	go func() {
		if err := Send(payload); err != nil {
			log.Printf("[Resend] Failed to send email to %v: %v", payload.To, err)
		}
	}()
}

func buildVerificationHTML(name, verifyLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;background:#f5f5f5;padding:32px">
<div style="max-width:480px;margin:0 auto;background:#fff;border-radius:12px;padding:32px">
  <h2 style="color:#0B1C3D;margin-top:0">Verifikasi Email Kamu</h2>
  <p style="color:#444">Halo %s,</p>
  <p style="color:#444">Klik tombol di bawah untuk memverifikasi alamat email kamu dan mulai menggunakan BengkelHub.</p>
  <a href="%s" style="display:inline-block;background:#E63946;color:#fff;text-decoration:none;padding:12px 28px;border-radius:8px;font-weight:bold;margin:16px 0">
    Verifikasi Email
  </a>
  <p style="color:#888;font-size:12px">Link ini berlaku selama 24 jam. Kalau kamu tidak mendaftar di BengkelHub, abaikan email ini.</p>
  <hr style="border:none;border-top:1px solid #eee;margin:24px 0">
  <p style="color:#aaa;font-size:12px">© BengkelHub</p>
</div>
</body>
</html>`, name, verifyLink)
}
