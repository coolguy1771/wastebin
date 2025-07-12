# Wastebin Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=wastebin
BINARY_UNIX=$(BINARY_NAME)_unix

# Build information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build clean test test-coverage test-integration test-security test-performance \
        test-all lint security-scan deps tidy check format help run dev docker \
        docker-build docker-run release install benchmark

# Default target
all: clean deps test build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/wastebin

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v ./cmd/wastebin

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html
	rm -f benchmark.txt

# Run tests
test:
	WASTEBIN_LOCAL_DB=true WASTEBIN_LOG_LEVEL=ERROR $(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@./scripts/test-coverage.sh

# Run integration tests
test-integration:
	WASTEBIN_LOCAL_DB=true WASTEBIN_LOG_LEVEL=ERROR $(GOTEST) -v -tags=integration ./tests/...

# Run security tests
test-security:
	WASTEBIN_LOCAL_DB=true WASTEBIN_LOG_LEVEL=ERROR $(GOTEST) -v ./tests/ -run TestSecurity

# Run performance tests
test-performance:
	WASTEBIN_LOCAL_DB=true WASTEBIN_LOG_LEVEL=ERROR $(GOTEST) -v ./tests/ -run TestPerformance -timeout 5m

# Run all tests
test-all: test test-integration test-security test-performance

# Run benchmarks
benchmark:
	WASTEBIN_LOCAL_DB=true WASTEBIN_LOG_LEVEL=ERROR $(GOTEST) -bench=. -benchmem -run=^$ ./... | tee benchmark.txt

# Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
		exit 1; \
	fi

# Run security scan
security-scan:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

# Install dependencies
deps:
	$(GOGET) -v ./...

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Check for issues
check: lint security-scan test

# Format code
format:
	gofmt -s -w .
	goimports -w .

# Run the application in development mode
dev:
	WASTEBIN_DEV=true WASTEBIN_LOCAL_DB=true $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/wastebin && ./$(BINARY_NAME)

# Run the application
run: build
	./$(BINARY_NAME)

# Docker build
docker-build:
	docker build -t wastebin:latest .

# Docker run
docker-run:
	docker run --rm -p 3000:3000 -e WASTEBIN_LOCAL_DB=true wastebin:latest

# Build and run with Docker
docker: docker-build docker-run

# Install binary to GOPATH/bin
install:
	$(GOCMD) install $(LDFLAGS) ./cmd/wastebin

# Create release build
release: clean
	@mkdir -p dist
	# Build for multiple platforms
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/wastebin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/wastebin
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/wastebin
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/wastebin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/wastebin

# Frontend commands
frontend-deps:
	cd web && npm ci

frontend-build:
	cd web && npm run build

frontend-dev:
	cd web && npm run dev

frontend-lint:
	cd web && npm run lint

# Full development setup
setup: deps frontend-deps

# Full build including frontend
build-all: frontend-build build

# Help
help:
	@echo "Available targets:"
	@echo "  build              Build the binary"
	@echo "  build-linux        Build for Linux"
	@echo "  clean              Clean build artifacts"
	@echo "  test               Run unit tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-integration   Run integration tests"
	@echo "  test-security      Run security tests"
	@echo "  test-performance   Run performance tests"
	@echo "  test-all           Run all tests"
	@echo "  benchmark          Run benchmarks"
	@echo "  lint               Run linter"
	@echo "  security-scan      Run security scan"
	@echo "  deps               Install dependencies"
	@echo "  tidy               Tidy dependencies"
	@echo "  check              Run linter, security scan, and tests"
	@echo "  format             Format code"
	@echo "  dev                Run in development mode"
	@echo "  run                Build and run"
	@echo "  docker-build       Build Docker image"
	@echo "  docker-run         Run Docker container"
	@echo "  docker             Build and run with Docker"
	@echo "  install            Install binary to GOPATH/bin"
	@echo "  release            Create release builds for multiple platforms"
	@echo "  frontend-deps      Install frontend dependencies"
	@echo "  frontend-build     Build frontend"
	@echo "  frontend-dev       Run frontend in development mode"
	@echo "  frontend-lint      Lint frontend code"
	@echo "  setup              Full development setup"
	@echo "  build-all          Build frontend and backend"
	@echo "  help               Show this help message"