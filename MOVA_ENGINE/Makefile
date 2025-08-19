# MOVA Engine - Makefile for Release Candidate v1.0.0
# =====================================================

# Variables
VERSION ?= 1.0.0-rc1
REGISTRY ?= ghcr.io/mova-engine
PROJECT_NAME = mova-engine
BINARY_NAME = mova
API_BINARY = mova-api

# Go build variables
GO_VERSION = 1.23.0
LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Docker variables
DOCKERFILE_API = infra/docker/Dockerfile
DOCKERFILE_CONSOLE = ../mova-console/Dockerfile

# Directories
BUILD_DIR = bin
DIST_DIR = dist
COVERAGE_DIR = coverage

.PHONY: help build test lint clean docker release install deps

# Default target
all: clean deps test lint build

# Help target
help: ## Show this help message
	@echo "MOVA Engine v${VERSION} - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION     - Release version (default: ${VERSION})"
	@echo "  REGISTRY    - Docker registry (default: ${REGISTRY})"

# =====================================================
# Dependencies
# =====================================================

deps: ## Install all dependencies
	@echo "📦 Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "📦 Installing Python SDK dependencies..."
	cd sdk/python && pip install -e . && pip install responses pytest pytest-cov
	@echo "📦 Installing TypeScript SDK dependencies..."
	cd sdk/typescript && npm install
	@echo "📦 Installing Console dependencies..."
	cd ../mova-console && npm install

# =====================================================
# Build
# =====================================================

build: ## Build all binaries
	@echo "🔨 Building MOVA Engine..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building CLI..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cli
	
	@echo "Building API Server..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(API_BINARY) ./api
	
	@echo "Building Console..."
	cd ../mova-console && npm run build
	
	@echo "✅ Build completed!"

build-cli: ## Build CLI binary only
	@echo "🔨 Building CLI..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cli

build-api: ## Build API server only
	@echo "🔨 Building API server..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(API_BINARY) ./api

build-console: ## Build Console only
	@echo "🔨 Building Console..."
	cd ../mova-console && npm run build

# Multi-platform builds
build-all-platforms: ## Build for all platforms
	@echo "🔨 Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext ./cli; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(DIST_DIR)/$(API_BINARY)-$$os-$$arch$$ext ./api; \
	done
	
	@echo "✅ Multi-platform build completed!"

# =====================================================
# Testing
# =====================================================

test: ## Run all tests
	@echo "🧪 Running all tests..."
	
	@echo "Running Go tests..."
	go test ./... -v -race -coverprofile=coverage.out
	
	@echo "Running Python SDK tests..."
	cd sdk/python && PYTHONPATH=./src python -m pytest tests/ -v --cov=src/mova --cov-report=html --cov-report=term
	
	@echo "Running TypeScript SDK tests..."
	cd sdk/typescript && npm test
	
	@echo "Running Console tests..."
	cd ../mova-console && npm test -- --watchAll=false --coverage
	
	@echo "✅ All tests completed!"

test-go: ## Run Go tests only
	@echo "🧪 Running Go tests..."
	go test ./... -v -race -coverprofile=coverage.out

test-python: ## Run Python SDK tests only
	@echo "🧪 Running Python SDK tests..."
	cd sdk/python && PYTHONPATH=./src python -m pytest tests/ -v --cov=src/mova

test-typescript: ## Run TypeScript SDK tests only
	@echo "🧪 Running TypeScript SDK tests..."
	cd sdk/typescript && npm test

test-console: ## Run Console tests only
	@echo "🧪 Running Console tests..."
	cd ../mova-console && npm test -- --watchAll=false

# Test coverage
coverage: ## Generate test coverage report
	@echo "📊 Generating coverage reports..."
	@mkdir -p $(COVERAGE_DIR)
	
	@echo "Go coverage..."
	go test ./... -coverprofile=$(COVERAGE_DIR)/go-coverage.out
	go tool cover -html=$(COVERAGE_DIR)/go-coverage.out -o $(COVERAGE_DIR)/go-coverage.html
	
	@echo "Python coverage..."
	cd sdk/python && PYTHONPATH=./src python -m pytest tests/ --cov=src/mova --cov-report=html:../../$(COVERAGE_DIR)/python-coverage
	
	@echo "TypeScript coverage..."
	cd sdk/typescript && npm run test -- --coverage --coverageDirectory=../../$(COVERAGE_DIR)/typescript-coverage
	
	@echo "Console coverage..."
	cd ../mova-console && npm test -- --coverage --coverageDirectory=../MOVA_ENGINE/$(COVERAGE_DIR)/console-coverage --watchAll=false
	
	@echo "✅ Coverage reports generated in $(COVERAGE_DIR)/"

# =====================================================
# Linting & Formatting
# =====================================================

lint: ## Run all linters
	@echo "🔍 Running linters..."
	
	@echo "Go linting..."
	golangci-lint run ./...
	
	@echo "Python linting..."
	cd sdk/python && flake8 src/ tests/
	cd sdk/python && black --check src/ tests/
	
	@echo "TypeScript linting..."
	cd sdk/typescript && npm run lint
	
	@echo "Console linting..."
	cd ../mova-console && npm run lint
	
	@echo "✅ All linting completed!"

format: ## Format all code
	@echo "🎨 Formatting code..."
	
	@echo "Go formatting..."
	go fmt ./...
	goimports -w .
	
	@echo "Python formatting..."
	cd sdk/python && black src/ tests/
	cd sdk/python && isort src/ tests/
	
	@echo "TypeScript formatting..."
	cd sdk/typescript && npm run format
	
	@echo "Console formatting..."
	cd ../mova-console && npm run format
	
	@echo "✅ All formatting completed!"

# =====================================================
# Security
# =====================================================

security: ## Run security scans
	@echo "🔒 Running security scans..."
	
	@echo "Go security scan..."
	gosec ./...
	
	@echo "Python security scan..."
	cd sdk/python && bandit -r src/
	
	@echo "NPM audit..."
	cd sdk/typescript && npm audit --audit-level=moderate
	cd ../mova-console && npm audit --audit-level=moderate
	
	@echo "✅ Security scans completed!"

# =====================================================
# Docker
# =====================================================

docker: ## Build Docker images
	@echo "🐳 Building Docker images..."
	
	@echo "Building API image..."
	docker build -t $(REGISTRY)/$(PROJECT_NAME)-api:$(VERSION) -f $(DOCKERFILE_API) .
	docker tag $(REGISTRY)/$(PROJECT_NAME)-api:$(VERSION) $(REGISTRY)/$(PROJECT_NAME)-api:latest
	
	@echo "Building Console image..."
	cd ../mova-console && docker build -t $(REGISTRY)/$(PROJECT_NAME)-console:$(VERSION) -f $(DOCKERFILE_CONSOLE) .
	cd ../mova-console && docker tag $(REGISTRY)/$(PROJECT_NAME)-console:$(VERSION) $(REGISTRY)/$(PROJECT_NAME)-console:latest
	
	@echo "✅ Docker images built!"

docker-push: ## Push Docker images to registry
	@echo "🚀 Pushing Docker images..."
	docker push $(REGISTRY)/$(PROJECT_NAME)-api:$(VERSION)
	docker push $(REGISTRY)/$(PROJECT_NAME)-api:latest
	docker push $(REGISTRY)/$(PROJECT_NAME)-console:$(VERSION)
	docker push $(REGISTRY)/$(PROJECT_NAME)-console:latest
	@echo "✅ Docker images pushed!"

# =====================================================
# Release
# =====================================================

release: ## Create release artifacts
	@echo "🚀 Creating release artifacts..."
	@mkdir -p $(DIST_DIR)
	
	# Build all platforms
	$(MAKE) build-all-platforms
	
	# Build Docker images
	$(MAKE) docker
	
	# Package Python SDK
	@echo "Packaging Python SDK..."
	cd sdk/python && python -m build --wheel --outdir ../../$(DIST_DIR)/
	
	# Package TypeScript SDK
	@echo "Packaging TypeScript SDK..."
	cd sdk/typescript && npm pack --pack-destination ../../$(DIST_DIR)/
	
	# Create checksums
	@echo "Creating checksums..."
	cd $(DIST_DIR) && sha256sum * > checksums.txt
	
	@echo "✅ Release artifacts created in $(DIST_DIR)/"

# =====================================================
# Installation
# =====================================================

install: build ## Install CLI binary to system
	@echo "📥 Installing MOVA CLI..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ MOVA CLI installed to /usr/local/bin/$(BINARY_NAME)"

uninstall: ## Uninstall CLI binary from system
	@echo "🗑️  Uninstalling MOVA CLI..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ MOVA CLI uninstalled"

# =====================================================
# Development
# =====================================================

dev-setup: deps ## Set up development environment
	@echo "🛠️  Setting up development environment..."
	
	# Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	
	# Python development tools
	pip install black isort flake8 bandit
	
	@echo "✅ Development environment ready!"

dev-api: ## Run API server in development mode
	@echo "🚀 Starting API server in development mode..."
	go run ./api --port=8080 --debug

dev-console: ## Run Console in development mode
	@echo "🚀 Starting Console in development mode..."
	cd ../mova-console && npm run dev

# =====================================================
# Utilities
# =====================================================

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f coverage.out
	cd sdk/python && rm -rf build/ dist/ *.egg-info/
	cd sdk/typescript && rm -rf dist/ node_modules/.cache/
	cd ../mova-console && rm -rf .next/ out/
	@echo "✅ Clean completed!"

validate: ## Validate project structure and dependencies
	@echo "✅ Validating project..."
	
	@echo "Checking Go modules..."
	go mod verify
	
	@echo "Checking Python setup..."
	cd sdk/python && python -m pip check
	
	@echo "Checking TypeScript setup..."
	cd sdk/typescript && npm audit --audit-level=high
	
	@echo "Checking Console setup..."
	cd ../mova-console && npm audit --audit-level=high
	
	@echo "✅ Validation completed!"

version: ## Show version information
	@echo "MOVA Engine $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "Node version: $(shell node --version)"
	@echo "Python version: $(shell python --version)"

# =====================================================
# Quick Commands
# =====================================================

quick-test: ## Quick test (Go only)
	go test ./... -short

quick-build: ## Quick build (CLI only)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cli

quick-lint: ## Quick lint (Go only)
	golangci-lint run ./... --fast