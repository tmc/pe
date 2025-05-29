.PHONY: build test scripttest clean

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