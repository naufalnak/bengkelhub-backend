.PHONY: run build tidy migrate-up migrate-down

# Run API server
run:
	go run ./cmd/api/main.go

# Run worker
run-worker:
	go run ./cmd/worker/main.go

# Build binary
build:
	go build -o bin/api ./cmd/api/main.go

# Tidy dependencies
tidy:
	go mod tidy

# Migrate up (requires golang-migrate CLI)
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

# Migrate down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down
