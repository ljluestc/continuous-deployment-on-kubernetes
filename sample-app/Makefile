# Makefile for Kubernetes Continuous Deployment Project

.PHONY: help test test-unit test-integration test-benchmark test-security test-performance test-all clean build run

# Default target
help:
	@echo "Available targets:"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-benchmark   - Run benchmark tests"
	@echo "  test-security    - Run security tests"
	@echo "  test-performance - Run performance tests"
	@echo "  test-all         - Run all tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  test-race        - Run tests with race detection"
	@echo "  test-short       - Run short tests only"
	@echo "  build            - Build the application"
	@echo "  run              - Run the application"
	@echo "  clean            - Clean build artifacts"
	@echo "  lint             - Run linting"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  mod-tidy         - Tidy go modules"

# Unit tests
test-unit:
	@echo "Running unit tests..."
	go test -tags=unit -v ./...

# Integration tests
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration -v ./...

# Benchmark tests
test-benchmark:
	@echo "Running benchmark tests..."
	go test -run=^$$ -bench=. -benchmem ./...

# Security tests
test-security:
	@echo "Running security tests..."
	go test -tags=security -v ./...

# Performance tests
test-performance:
	@echo "Running performance tests..."
	go test -tags=performance -v ./...

# All tests
test-all: test-unit test-integration test-benchmark test-security test-performance
	@echo "All tests completed!"

# Tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./...

# Short tests only
test-short:
	@echo "Running short tests..."
	go test -short -v ./...

# Build the application
build:
	@echo "Building application..."
	go build -v -ldflags="-s -w" -o sample-app .

# Run the application
run: build
	@echo "Running application..."
	./sample-app

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f sample-app
	rm -f coverage.out
	rm -f coverage.html
	rm -f test_results.json
	rm -f test_report.md

# Linting
lint:
	@echo "Running linters..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Tidy go modules
mod-tidy:
	@echo "Tidying go modules..."
	go mod tidy

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod vendor

# Run comprehensive test suite
test-comprehensive:
	@echo "Running comprehensive test suite..."
	python3 ../test_comprehensive.py --project-root .. --output test_results.json --report test_report.md

# Run pre-commit hooks
pre-commit:
	@echo "Running pre-commit hooks..."
	pre-commit run --all-files

# Install pre-commit hooks
install-hooks:
	@echo "Installing pre-commit hooks..."
	pre-commit install

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t sample-app .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 sample-app

# Kubernetes deploy
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/

# Kubernetes delete
k8s-delete:
	@echo "Deleting from Kubernetes..."
	kubectl delete -f k8s/

# Show help
.DEFAULT_GOAL := help
