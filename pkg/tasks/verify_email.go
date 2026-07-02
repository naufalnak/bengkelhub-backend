package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/naufalnak/bengkelhub-backend/pkg/resend"
)

const TypeSendVerificationEmail = "email:verify"

type VerificationEmailPayload struct {
	ToEmail    string `json:"to_email"`
	ToName     string `json:"to_name"`
	VerifyLink string `json:"verify_link"`
}

func NewVerificationEmailTask(payload VerificationEmailPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendVerificationEmail, data), nil
}

// HandleVerificationEmailTask dijalankan di worker
func HandleVerificationEmailTask(t *asynq.Task) error {
	var p VerificationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}
	return resend.SendVerificationEmail(p.ToEmail, p.ToName, p.VerifyLink)
}
