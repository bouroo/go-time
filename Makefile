# Go Project Makefile
# Comprehensive build and development automation for the go-time library

# =============================================================================
# Variables
# =============================================================================

# Project Information
MODULE_NAME := $(shell go list -m)
BINARY_NAME := go-time

# Go Flags
GOFLAGS ?= -mod=mod
GO_BUILD_FLAGS ?= -v

# Directories
BUILD_DIR ?= ./bin
COVERAGE_DIR ?= ./coverage

# Test Flags
TEST_FLAGS ?= -v -race
BENCH_FLAGS ?= -bench=. -benchmem
COVERPROFILE ?= coverage.out

# Colors for terminal output (disabled on non-TTY)
ifneq ($(shell test -t 0 && echo true),true)
  GREEN :=
  YELLOW :=
  BLUE :=
  NC :=
else
  GREEN := \033[0;32m
  YELLOW := \033[0;33m
  BLUE := \033[0;34m
  NC := \033[0m
endif

# =============================================================================
# Build Targets
# =============================================================================

.PHONY: build clean

# Build the project binary
# Usage: make build
build: $(BUILD_DIR)
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	go build $(GOFLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Clean build artifacts
# Usage: make clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR) $(COVERPROFILE)
	@echo "$(GREEN)Clean complete$(NC)"

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# =============================================================================
# Test Targets
# =============================================================================

.PHONY: test-unit test-all test-bench test-coverage test-race test

# Run unit tests (default test target)
# Usage: make test-unit
test-unit:
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test $(TEST_FLAGS) ./...
	@echo "$(GREEN)Unit tests passed$(NC)"

# Run all tests with coverage
# Usage: make test-all
test-all: $(COVERAGE_DIR)
	@echo "$(BLUE)Running all tests with coverage...$(NC)"
	go test $(TEST_FLAGS) -coverprofile=$(COVERAGE_DIR)/$(COVERPROFILE) -covermode=atomic ./...
	@echo "$(GREEN)All tests passed$(NC)"

# Run benchmarks
# Usage: make test-bench
test-bench:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	go test $(BENCH_FLAGS) ./...
	@echo "$(GREEN)Benchmarks complete$(NC)"

# Generate coverage report with HTML output
# Usage: make test-coverage
test-coverage: $(COVERAGE_DIR)
	@echo "$(BLUE)Generating coverage report...$(NC)"
	go test $(TEST_FLAGS) -coverprofile=$(COVERAGE_DIR)/$(COVERPROFILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_DIR)/$(COVERPROFILE) -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

# Run tests with race detector only (separate from unit tests)
# Usage: make test-race
test-race:
	@echo "$(BLUE)Running race detector...$(NC)"
	go test -race ./...
	@echo "$(GREEN)Race detector passed$(NC)"

# Run all tests (convenience target)
# Usage: make test
test: test-unit
	@echo "$(GREEN)All tests passed$(NC)"

$(COVERAGE_DIR):
	mkdir -p $(COVERAGE_DIR)

# =============================================================================
# Code Quality Targets
# =============================================================================

.PHONY: fmt lint vet check check-diff

# Format code with gofmt
# Usage: make fmt
fmt:
	@echo "$(BLUE)Formatting code with gofmt...$(NC)"
	gofmt -w .
	@echo "$(GREEN)Code formatted$(NC)"

# Run golangci-lint
# Usage: make lint
lint:
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)Lint passed$(NC)"

# Run go vet
# Usage: make vet
vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)Go vet passed$(NC)"

# Run all code quality checks
# Usage: make check
check: fmt vet lint
	@echo "$(GREEN)All code quality checks passed$(NC)"

# Check for formatting differences (useful for CI)
# Usage: make check-diff
check-diff:
	@echo "$(BLUE)Checking code formatting...$(NC)"
	@gofmt -d . > /dev/null 2>&1 && echo "Code is properly formatted" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)

# =============================================================================
# Dependency Management
# =============================================================================

.PHONY: deps tidy verify update

# Download dependencies
# Usage: make deps
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	@echo "$(GREEN)Dependencies downloaded$(NC)"

# Tidy go.mod and go.sum
# Usage: make tidy
tidy:
	@echo "$(BLUE)Running go mod tidy...$(NC)"
	go mod tidy
	@echo "$(GREEN)Module tidy complete$(NC)"

# Verify dependencies (check if go.sum matches go.mod)
# Usage: make verify
verify:
	@echo "$(BLUE)Verifying dependencies...$(NC)"
	go mod verify
	@echo "$(GREEN)Dependencies verified$(NC)"

# Update dependencies to latest versions
# Usage: make update
update:
	@echo "$(BLUE)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

# =============================================================================
# Development Helpers
# =============================================================================

.PHONY: install run help

# Install the package
# Usage: make install
install:
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	go install .
	@echo "$(GREEN)Installed $(BINARY_NAME)$(NC)"

# Run the main package (if applicable)
# Usage: make run
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

# Display help for all targets
# Usage: make help
help:
	@echo "Go Project Makefile - $(MODULE_NAME)"
	@echo ""
	@echo "Build Targets:"
	@echo "  build     - Build the project binary"
	@echo "  clean     - Clean build artifacts"
	@echo ""
	@echo "Test Targets:"
	@echo "  test          - Run unit tests"
	@echo "  test-unit     - Run unit tests"
	@echo "  test-all      - Run all tests with coverage"
	@echo "  test-bench    - Run benchmarks"
	@echo "  test-coverage - Generate coverage report with HTML"
	@echo "  test-race     - Run tests with race detector"
	@echo ""
	@echo "Code Quality Targets:"
	@echo "  fmt        - Format code with gofmt"
	@echo "  lint       - Run golangci-lint"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run all code quality checks (fmt, vet, lint)"
	@echo "  check-diff - Check for formatting differences"
	@echo ""
	@echo "Dependency Management:"
	@echo "  deps   - Download dependencies"
	@echo "  tidy   - Run go mod tidy"
	@echo "  verify - Verify dependencies"
	@echo "  update - Update dependencies to latest versions"
	@echo ""
	@echo "Development Helpers:"
	@echo "  install - Install the package"
	@echo "  run     - Build and run the project"
	@echo "  help    - Display this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME   - Name of the binary (default: go-time)"
	@echo "  GOFLAGS       - Go flags (default: -mod=mod)"
	@echo "  BUILD_DIR     - Build output directory (default: ./bin)"
	@echo "  COVERAGE_DIR  - Coverage output directory (default: ./coverage)"
	@echo "  COVERPROFILE  - Coverage profile name (default: coverage.out)"
