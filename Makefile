.PHONY: all build test lint fmt check up down migrate migrate-create clean install dev

BINARY_NAME=acumius
MAIN_PATH=./cmd/acumius
DOCKER_COMPOSE=docker-compose

all: check

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

install:
	go mod download
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

dev:
	go run $(MAIN_PATH)

dev-up:
	$(DOCKER_COMPOSE) up -d postgres valkey

up:
	$(DOCKER_COMPOSE) up -d

down:
	$(DOCKER_COMPOSE) down

down-volumes:
	$(DOCKER_COMPOSE) down -v

migrate:
	migrate -path ./migrations -database $(ACUMIUS_DATABASE_URL) up

migrate-down:
	migrate -path ./migrations -database $(ACUMIUS_DATABASE_URL) down

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

test:
	go test -v -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	go vet ./...
	golangci-lint run

fmt:
	go fmt ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

check: fmt-check lint test

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make dev            - Run in development mode"
	@echo "  make up             - Start all services with docker-compose"
	@echo "  make down           - Stop all services"
	@echo "  make migrate        - Run database migrations"
	@echo "  make migrate-create - Create a new migration (name=...)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code"
	@echo "  make check          - Run fmt-check, lint, and test"
	@echo "  make bench          - Run benchmarks"
	@echo "  make clean          - Clean build artifacts"
