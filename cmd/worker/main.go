package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/url"

	"github.com/hibiken/asynq"
	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/worker"
	"github.com/naufalnak/bengkelhub-backend/pkg/tasks"
)

type errorHandler struct{}

func (h *errorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	log.Printf("[Worker] Task failed: type=%s err=%v", task.Type(), err)
}

func main() {
	config.Load()

	if config.Cfg.RedisURL == "" {
		log.Fatal("[Worker] REDIS_URL is required")
	}

	opt, err := parseRedisOpt(config.Cfg.RedisURL)
	if err != nil {
		log.Fatalf("[Worker] Failed to parse REDIS_URL: %v", err)
	}

	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 5,
				"default":  10,
				"email":    8, // queue khusus email, prioritas lebih tinggi dari default
			},
			ErrorHandler: &errorHandler{},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeReminderBooking, worker.HandleReminderBooking)
	mux.HandleFunc(tasks.TypeSendVerificationEmail, tasks.HandleVerificationEmailTask)

	log.Println("[Worker] Starting Asynq worker...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("[Worker] Failed to start: %v", err)
	}
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