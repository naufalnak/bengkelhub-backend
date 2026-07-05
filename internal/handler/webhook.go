package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	mt "github.com/naufalnak/bengkelhub-backend/pkg/midtrans"
	"github.com/naufalnak/bengkelhub-backend/internal/service"
	"github.com/naufalnak/bengkelhub-backend/pkg/response"
)

type WebhookHandler struct {
	invoiceService service.InvoiceService
}

func NewWebhookHandler(invoiceService service.InvoiceService) *WebhookHandler {
	return &WebhookHandler{invoiceService}
}

func (h *WebhookHandler) HandlePayment(c *fiber.Ctx) error {
	var payload mt.WebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("[Webhook] Failed to parse body: %v", err)
		return response.Error(c, fiber.StatusBadRequest, "Invalid webhook payload", nil)
	}

	log.Printf("[Webhook] Received: order_id=%s status=%s payment_type=%s",
		payload.OrderID, payload.TransactionStatus, payload.PaymentType)

	if err := h.invoiceService.HandleWebhook(payload); err != nil {
		log.Printf("[Webhook] HandleWebhook error: %v", err)
		if err.Error() == "invalid webhook signature" {
			return response.Error(c, fiber.StatusBadRequest, "Invalid signature", nil)
		}
		log.Printf("[Webhook] Non-critical error, returning 200: %v", err)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}