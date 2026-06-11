package main

import (
	"log"

	"github.com/naufalnak/bengkelku-api/config"
)

func main() {
	config.Load()

	// TODO: Init Asynq worker di bulan 3
	log.Println("Worker placeholder — Asynq integration coming soon")
}