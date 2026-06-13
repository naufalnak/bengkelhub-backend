package tasks

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/hibiken/asynq"
	"github.com/naufalnak/bengkelhub-backend/config"
)

var client *asynq.Client

func InitClient() {
	if config.Cfg.RedisURL == "" {
		log.Println("[Asynq] REDIS_URL not set, task queue disabled")
		return
	}

	opt, err := parseRedisOpt(config.Cfg.RedisURL)
	if err != nil {
		log.Printf("[Asynq] Failed to parse REDIS_URL: %v", err)
		return
	}

	client = asynq.NewClient(opt)
	log.Println("[Asynq] Task queue client initialized")
}

func EnqueueReminderBooking(payload ReminderBookingPayload) error {
	if client == nil {
		log.Println("[Asynq] Client not initialized, skipping enqueue")
		return nil
	}

	task, err := NewReminderBookingTask(payload)
	if err != nil {
		return err
	}

	// Jadwalkan H-1 (24 jam sebelum slot)
	// Untuk test: ganti dengan time.Now().Add(10 * time.Second)
	runAt := payload.SlotDate.Add(-24 * time.Hour)
	if runAt.Before(time.Now()) {
		log.Printf("[Asynq] Reminder time already passed for order %s, skipping", payload.OrderID)
		return nil
	}

	info, err := client.Enqueue(task,
		asynq.ProcessAt(runAt),
		asynq.MaxRetry(3),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Printf("[Asynq] Reminder enqueued: id=%s queue=%s processAt=%s",
		info.ID, info.Queue, runAt.Format(time.RFC3339))
	return nil
}

func parseRedisOpt(rawURL string) (asynq.RedisClientOpt, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return asynq.RedisClientOpt{}, fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	opt := asynq.RedisClientOpt{
		Addr: u.Host,
		// Upstash requires TLS
		TLSConfig: &tls.Config{},
	}

	if u.User != nil {
		if pass, ok := u.User.Password(); ok {
			opt.Password = pass
		}
	}

	return opt, nil
}