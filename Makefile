.PHONY: build test scripttest clean install lint coverage

# Default target
all: build

# Build the pe binary
build:
	go build -o pe ./cmd/pe

# Run Go tests
test:
	go test ./...

# Run scripttest tests
scripttest: build
	@echo "Running scripttests..."
	@cd test && go run run-scripttest.go module-registry.scripttest

# Clean build artifacts
clean:
	rm -f pe
	rm -rf .pe/cache
	rm -rf .pe/modules
	rm -rf .pe/test

# Install pe binary
install: build
	go install ./cmd/pe

# Run specific scripttest
test-module: build
	cd test && go run run-scripttest.go module-basic.scripttest

# Setup mock registry for testing
mock-registry:
	./test/mock-registry.sh

# Run linting
lint:
	golangci-lint run ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run all checks (test, lint)
check: test lint

# Development setup
dev-setup:
	go mod download
	go mod verify
	@echo "Development environment ready"