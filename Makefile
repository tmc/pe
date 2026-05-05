.PHONY: build test scripttest clean install lint coverage coverage-check security bench check

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

# Check total coverage against COVERAGE_MIN.
coverage-check:
	@min=$${COVERAGE_MIN:-45}; \
	go test -coverprofile=coverage.out ./...; \
	total=$$(go tool cover -func=coverage.out | awk '/^total:/ { sub(/%/, "", $$3); print $$3 }'); \
	awk -v total="$$total" -v min="$$min" 'BEGIN { if (total+0 < min+0) { printf "coverage %.1f%% below %.1f%%\n", total, min; exit 1 } printf "coverage %.1f%% >= %.1f%%\n", total, min }'

# Run dependency and code security checks.
security:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Run benchmarks without producing artifacts.
bench:
	go test -run '^$$' -bench . -benchmem ./cmd/pe ./internal/...

# Run all checks (test, lint)
check: test coverage-check security

# Development setup
dev-setup:
	go mod download
	go mod verify
	@echo "Development environment ready"
