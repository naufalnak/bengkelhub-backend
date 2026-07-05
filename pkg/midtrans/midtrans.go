package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/naufalnak/bengkelhub-backend/config"
)

func baseURL() string {
	if config.Cfg.MidtransEnv == "production" {
		return "https://app.midtrans.com/snap/v1"
	}
	return "https://app.sandbox.midtrans.com/snap/v1"
}

type TransactionDetail struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type CustomerDetail struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type ItemDetail struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type CreateTransactionRequest struct {
	TransactionDetail TransactionDetail `json:"transaction_details"`
	CustomerDetail    CustomerDetail    `json:"customer_details,omitempty"`
	ItemDetails       []ItemDetail      `json:"item_details,omitempty"`
	Callbacks         *Callbacks        `json:"callbacks,omitempty"`
}

type Callbacks struct {
	Finish string `json:"finish"`
}

type CreateTransactionResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type WebhookPayload struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
}

func CreateTransaction(req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	if config.Cfg.MidtransServerKey == "" {
		return nil, fmt.Errorf("MIDTRANS_SERVER_KEY not configured")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest(http.MethodPost, baseURL()+"/transactions", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.SetBasicAuth(config.Cfg.MidtransServerKey, "")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("midtrans request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("midtrans error %d: %s", resp.StatusCode, string(raw))
	}

	var result CreateTransactionResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse midtrans response: %w", err)
	}

	log.Printf("[Midtrans] Transaction created: order_id=%s url=%s", req.TransactionDetail.OrderID, result.RedirectURL)
	return &result, nil
}

func VerifySignature(payload WebhookPayload) bool {
	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + config.Cfg.MidtransServerKey
	hash := sha512.Sum512([]byte(raw))
	computed := fmt.Sprintf("%x", hash)
	return strings.EqualFold(computed, payload.SignatureKey)
}

func IsPaymentSuccess(payload WebhookPayload) bool {
	switch payload.TransactionStatus {
	case "capture":
		return payload.FraudStatus == "accept"
	case "settlement":
		return true
	default:
		return false
	}
}

func PaymentMethodFromType(paymentType string) string {
	switch {
	case strings.Contains(paymentType, "qris") ||
		strings.Contains(paymentType, "gopay") ||
		strings.Contains(paymentType, "shopeepay"):
		return "qris"
	case paymentType == "credit_card":
		return "transfer"
	default:
		return "transfer"
	}
}