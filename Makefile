.PHONY: help build run stop clean test docker-build docker-up docker-down docker-logs migrate-up migrate-down sqlc-generate swagger-generate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application binary
	go build -o bin/chat ./cmd/chat

run: ## Run the application locally
	go run ./cmd/chat

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

docker-build: ## Build Docker image
	docker build -t chat-service:latest .

docker-up: ## Start services with docker-compose
	docker-compose up -d

docker-down: ## Stop services with docker-compose
	docker-compose down

docker-logs: ## View docker-compose logs
	docker-compose logs -f

docker-restart: docker-down docker-up ## Restart docker-compose services

sqlc-generate: ## Generate database code with sqlc
	sqlc generate

swagger-generate: ## Generate Swagger documentation
	swag init -g cmd/chat/main.go -o docs --parseDependency --parseInternal

migrate-up: ## Run database migrations (requires goose)
	goose -dir migrations postgres "$(DATABASE_DSN)" up

migrate-down: ## Rollback database migrations (requires goose)
	goose -dir migrations postgres "$(DATABASE_DSN)" down

dev: ## Start development environment
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "PostgreSQL is ready. You can now run 'make run' or 'make migrate-up'"

.DEFAULT_GOAL := help
