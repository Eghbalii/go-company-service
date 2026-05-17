BINARY_NAME=company-service
BUILD_DIR=./bin
MAIN_PATH=./cmd/api
MODULE=github.com/eghbalii/go-company-service
MIGRATE_PATH=./migrations
DB_URL?=postgres://postgres:postgres@localhost:5432/company_db?sslmode=disable

.PHONY: all build run test test-unit test-integration lint fmt swagger migrate-up migrate-down migrate-create docker-up docker-down clean tidy

all: lint test build

## Build
build:
	@echo "Building..."
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run:
	@go run $(MAIN_PATH)

## Dependencies
tidy:
	@go mod tidy

## Testing
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test -v -race -count=1 -coverprofile=coverage.out ./internal/usecase/... ./internal/adapter/http/...
	@go tool cover -func=coverage.out

test-integration:
	@echo "Running integration tests..."
	@go test -v -race -count=1 -tags=integration ./internal/integration/...

test-cover:
	@go test -v -race -count=1 -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Code quality
lint:
	@echo "Linting..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting..."
	@gofmt -s -w .
	@goimports -w .

vet:
	@go vet ./...

## Swagger
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

## Database migrations
migrate-up:
	@echo "Running migrations up..."
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" up

migrate-down:
	@echo "Running migrations down..."
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $$name

migrate-version:
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" version

migrate-force:
	@read -p "Version: " ver; \
	migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" force $$ver

## Docker
docker-up:
	@echo "Starting services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping services..."
	@docker-compose down

docker-build:
	@docker-compose build

docker-logs:
	@docker-compose logs -f app

## Cleanup
clean:
	@rm -rf $(BUILD_DIR) coverage.out coverage.html

## Install tools
install-tools:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.61.0

.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:' Makefile | grep -v '^\.' | awk -F: '{print "  " $$1}' | sort
