# Makefile for Server Management

APP_NAME=server-management
DOCKER_IMAGE=servermanagement:latest

.PHONY: help run build test lint swag docker-build docker-up docker-down docker-reload air-local air-docker

help:
	@echo "Available targets:"
	@echo "  run           - Run the app locally (go run)"
	@echo "  build         - Build the Go binary"
	@echo "  test          - Run tests with race detector"
	@echo "  lint          - Run golangci-lint if available"
	@echo "  swag          - Generate Swagger docs"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-up     - Start app and Postgres with Docker Compose"
	@echo "  docker-down   - Stop Docker Compose"
	@echo "  docker-reload - Autoreload: build, docker-compose up, restart on save (needs air, Docker)"
	@echo "  air-local     - Local Go hot-reload with Air (no Docker)"
	@echo "  air-docker    - Docker Compose hot-reload with Air (uses Docker)"

run:
	go run ./cmd/main.go

build:
	go build -o $(APP_NAME) ./cmd/main.go

test:
	go test -race -count=1 ./...

lint:
	@if [ -x "$(shell command -v golangci-lint 2>/dev/null)" ]; then \
		golangci-lint run ./... ; \
	else \
		echo "golangci-lint not installed" ; \
	fi

swag:
	swag init -g cmd/main.go

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

# Autoreload: requires 'air' (https://github.com/air-verse/air) and Docker Compose
# On file change, rebuilds and restarts the container
# You can install air with: go install github.com/cosmtrek/air@latest

docker-reload:
	@echo "Starting Docker Compose autoreload with air (requires air installed)"
	air -c .air.docker.toml || echo "Install air: go install github.com/cosmtrek/air@latest"

air-local:
	@echo "Starting local Go hot-reload with air (requires air installed)"
	air -c .air.local.toml || echo "Install air: go install github.com/cosmtrek/air@latest"

air-docker:
	@echo "Starting Docker Compose hot-reload with air (requires air installed)"
	air -c .air.docker.toml || echo "Install air: go install github.com/cosmtrek/air@latest"
