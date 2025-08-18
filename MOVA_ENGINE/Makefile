.PHONY: help build test clean docker-build docker-run dev-setup

# Default target
help:
	@echo "MOVA Engine - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  build         - Build all components"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  dev-setup     - Setup development environment"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-stop   - Stop Docker services"
	@echo ""
	@echo "Individual components:"
	@echo "  build-core    - Build Go core engine"
	@echo "  build-api     - Build API server"
	@echo "  build-cli     - Build CLI tool"
	@echo "  build-sdk-ts  - Build TypeScript SDK"
	@echo "  build-sdk-py  - Build Python SDK"
	@echo "  build-web     - Build web console"

# Build all components
build: build-core build-api build-cli build-sdk-ts build-sdk-py build-web

# Build Go core engine
build-core:
	@echo "Building Go core engine..."
	cd core && go build -o ../bin/mova-core ./...

# Build API server
build-api:
	@echo "Building API server..."
	cd api && go build -o ../bin/mova-api .

# Build CLI tool
build-cli:
	@echo "Building CLI tool..."
	cd cli && go build -o ../bin/mova .

# Build TypeScript SDK
build-sdk-ts:
	@echo "Building TypeScript SDK..."
	cd sdk/typescript && npm ci && npm run build

# Build Python SDK
build-sdk-py:
	@echo "Building Python SDK..."
	cd sdk/python && pip install -e ".[dev]"

# Build web console
build-web:
	@echo "Building web console..."
	cd web && npm ci && npm run build

# Run all tests
test:
	@echo "Running Go tests..."
	cd core && go test -v ./...
	cd api && go test -v ./...
	cd cli && go test -v ./...
	@echo "Running TypeScript tests..."
	cd sdk/typescript && npm test
	@echo "Running Python tests..."
	cd sdk/python && pytest

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf sdk/typescript/dist/
	rm -rf sdk/typescript/node_modules/
	rm -rf web/.next/
	rm -rf web/node_modules/
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

# Setup development environment
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing Go dependencies..."
	cd core && go mod download
	cd api && go mod download
	cd cli && go mod download
	@echo "Installing Node.js dependencies..."
	cd sdk/typescript && npm ci
	cd web && npm ci
	@echo "Installing Python dependencies..."
	cd sdk/python && pip install -e ".[dev]"
	@echo "Development environment setup complete!"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -f infra/docker/Dockerfile -t mova-engine:latest .

docker-run:
	@echo "Starting services with Docker Compose..."
	cd infra/docker && docker-compose up -d

docker-stop:
	@echo "Stopping Docker services..."
	cd infra/docker && docker-compose down

# Development server commands
dev-api:
	@echo "Starting API server in development mode..."
	cd api && go run main.go

dev-web:
	@echo "Starting web console in development mode..."
	cd web && npm run dev

# Linting and formatting
lint:
	@echo "Running linters..."
	@echo "Go linting..."
	cd core && golint ./...
	cd api && golint ./...
	cd cli && golint ./...
	@echo "TypeScript linting..."
	cd sdk/typescript && npm run lint
	@echo "Python linting..."
	cd sdk/python && flake8 mova_engine_sdk
	cd sdk/python && black --check mova_engine_sdk

format:
	@echo "Formatting code..."
	@echo "Go formatting..."
	cd core && go fmt ./...
	cd api && go fmt ./...
	cd cli && go fmt ./...
	@echo "TypeScript formatting..."
	cd sdk/typescript && npm run format
	@echo "Python formatting..."
	cd sdk/python && black mova_engine_sdk

# Database commands
db-migrate:
	@echo "Running database migrations..."
	@echo "Not implemented yet"

db-seed:
	@echo "Seeding database..."
	@echo "Not implemented yet"

# Monitoring commands
monitor:
	@echo "Starting monitoring stack..."
	cd infra/docker && docker-compose up -d prometheus grafana
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

# Documentation
docs:
	@echo "Generating documentation..."
	@echo "Not implemented yet"

# Release
release:
	@echo "Creating release..."
	@echo "Not implemented yet"
