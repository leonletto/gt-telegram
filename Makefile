.PHONY: help test build install clean fmt vet ci quick-check

# Binary name and install location
BINARY_NAME := gt-telegram
BUILD_DIR := bin
INSTALL_DIR := $(HOME)/.local/bin
VERSION := 0.1.0

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  make build        - Build binary"
	@echo "  make install      - Build and install to ~/.local/bin"
	@echo "  make fmt          - Format Go code"
	@echo "  make vet          - Run go vet"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make tidy         - Tidy dependencies"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run all tests"
	@echo "  make test-race    - Run tests with race detector"
	@echo ""
	@echo "Quality:"
	@echo "  make quick-check  - Fast pre-commit checks (fmt, vet, test, build)"
	@echo "  make ci           - Full CI checks locally"
	@echo ""

# Build binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-X main.Version=$(VERSION) -X main.Build=$$(git rev-parse --short HEAD)" -o $(BUILD_DIR)/$(BINARY_NAME) .
ifeq ($(shell uname),Darwin)
	@codesign -s - -f $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || true
	@echo "Signed $(BINARY_NAME) for macOS"
endif
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary to ~/.local/bin
install: build
	@mkdir -p $(INSTALL_DIR)
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"

# Run all tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test ./... -v -race

# Format Go code
fmt:
	@echo "Formatting Go code..."
	gofmt -s -w -e .

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)/
	rm -rf dist/
	go clean

# Quick pre-commit checks
quick-check: fmt vet test build
	@echo "Quick checks passed"

# Full CI checks locally
ci: fmt vet test-race build
	@echo "CI checks passed"
