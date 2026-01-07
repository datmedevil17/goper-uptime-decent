.PHONY: help install build run test clean migrate docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make install      - Install dependencies"
	@echo "  make build        - Build all binaries"
	@echo "  make run-api      - Run API server"
	@echo "  make run-hub      - Run Hub server"
	@echo "  make run-validator- Run Validator"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate      - Run database migrations"
	@echo "  make docker-up    - Start all services"
	@echo "  make docker-down  - Stop all services"
	@echo "  make docker-logs  - View logs"
	@echo "  make db-up        - Start database and rabbitmq"
	@echo "  make db-down      - Stop database and rabbitmq"
	@echo "  make db-shell     - Open database shell"

install:
	go mod download
	go mod tidy

build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/hub ./cmd/hub
	go build -o bin/validator ./cmd/validator
	@echo "✅ Build complete"

run-api:
	go run ./cmd/api/main.go

run-hub:
	go run ./cmd/hub/main.go

run-validator:
	PRIVATE_KEY=${VALIDATOR_PRIVATE_KEY} go run ./cmd/validator/main.go

test:
	go test -v -cover ./...

clean:
	rm -rf bin/
	go clean

migrate:
	@echo "Running migrations..."
	go run ./cmd/api/main.go

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

generate-keypair:
	@echo "Generating Solana keypair..."
	solana-keygen new --outfile validator-keypair.json
	@echo "✅ Keypair saved to validator-keypair.json"
	@echo "Public key:"
	@solana-keygen pubkey validator-keypair.json

db-up:
	docker compose up -d postgres rabbitmq

db-down:
	docker compose stop postgres rabbitmq

db-shell:
	docker compose exec postgres psql -U uptime_user -d uptime_db