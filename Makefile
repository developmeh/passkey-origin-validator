# Makefile for passkey-origin-validator

# Variables
BINARY_NAME=passkey-origin-validator
CMD_DIR=cmd/passkey-origin-validator
BUILD_DIR=build
DEBUG=false
DOMAIN=
FILE=
ORIGIN=

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@CMD="./$(BUILD_DIR)/$(BINARY_NAME) count"; \
	if [ "$(DEBUG)" = "true" ]; then \
		echo "Debug mode enabled"; \
		CMD="$$CMD --debug"; \
	fi; \
	if [ -n "$(FILE)" ]; then \
		echo "Using file: $(FILE)"; \
		CMD="$$CMD --file $(FILE)"; \
	elif [ -n "$(DOMAIN)" ]; then \
		echo "Testing domain: $(DOMAIN)"; \
		CMD="$$CMD $(DOMAIN)"; \
	fi; \
	eval $$CMD

# Run the validate command
.PHONY: validate
validate: build
	@echo "Running $(BINARY_NAME) validate..."
	@if [ -z "$(ORIGIN)" ]; then \
		echo "Error: ORIGIN is required for validate command"; \
		echo "Usage: make validate ORIGIN=https://example.com [DOMAIN=domain.com] [FILE=file.json] [DEBUG=true]"; \
		exit 1; \
	fi; \
	CMD="./$(BUILD_DIR)/$(BINARY_NAME) validate --origin $(ORIGIN)"; \
	if [ "$(DEBUG)" = "true" ]; then \
		echo "Debug mode enabled"; \
		CMD="$$CMD --debug"; \
	fi; \
	if [ -n "$(FILE)" ]; then \
		echo "Using file: $(FILE)"; \
		CMD="$$CMD --file $(FILE)"; \
	elif [ -n "$(DOMAIN)" ]; then \
		echo "Testing domain: $(DOMAIN)"; \
		CMD="$$CMD $(DOMAIN)"; \
	fi; \
	eval $$CMD

# Run with mock data
.PHONY: mock
mock: build
	@echo "Running $(BINARY_NAME) with mock data..."
	@CMD="./$(BUILD_DIR)/$(BINARY_NAME) --mock"; \
	if [ "$(DEBUG)" = "true" ]; then \
		echo "Debug mode enabled"; \
		CMD="$$CMD --debug"; \
	fi; \
	eval $$CMD

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Get dependencies
.PHONY: deps
deps:
	@echo "Getting dependencies..."
	@go mod tidy

# Help
.PHONY: help
help:
	@echo "Makefile for $(BINARY_NAME)"
	@echo ""
	@echo "Usage:"
	@echo "  make [target] [options]"
	@echo ""
	@echo "Targets:"
	@echo "  all       Build the application (default)"
	@echo "  build     Build the application"
	@echo "  run       Build and run the count command"
	@echo "  validate  Build and run the validate command"
	@echo "  mock      Build and run with mock data"
	@echo "  clean     Remove build artifacts"
	@echo "  test      Run tests"
	@echo "  deps      Get dependencies"
	@echo "  help      Show this help message"
	@echo ""
	@echo "Options:"
	@echo "  DEBUG     Enable debug mode (default: false)"
	@echo "  DOMAIN    Specify a domain to test (default: webauthn.io)"
	@echo "  FILE      Use a local JSON file instead of fetching from a domain"
	@echo "  ORIGIN    The caller origin to validate (required for validate command)"
	@echo ""
	@echo "Examples:"
	@echo "  make run                                  # Count labels for default domain"
	@echo "  make run DEBUG=true                       # Count labels with debug logging"
	@echo "  make run DOMAIN=google.com                # Count labels for specific domain"
	@echo "  make run FILE=./test.json                 # Count labels from local file"
	@echo "  make validate ORIGIN=https://example.com  # Validate origin against default domain"
	@echo "  make validate ORIGIN=https://example.com DOMAIN=google.com  # Validate against specific domain"
	@echo "  make validate ORIGIN=https://example.com FILE=./test.json   # Validate against local file"
	@echo "  make mock                                 # Run with mock data"
