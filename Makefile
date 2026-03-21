.PHONY: all setup build build-all test test-unit test-integration coverage fmt lint vet check run clean help

# Variables
BINARY_NAME=mcp-obsidian-go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=build
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"

# Default target
all: build

## Setup development environment
setup:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Build local binary
build:
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

## Build for all platforms
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 build-windows-amd64

build-darwin-amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server

build-darwin-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server

build-linux-amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server

build-linux-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/server

build-windows-amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

## Run tests
test:
	go test -v -race ./...

## Run unit tests only
test-unit:
	go test -v -short ./...

## Run integration tests
test-integration:
	go test -v -tags=integration ./...

## Generate coverage report
coverage:
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

## Format code
fmt:
	go fmt ./...

## Run linter
lint:
	golangci-lint run

## Run go vet
vet:
	go vet ./...

## Run all checks (requires golangci-lint: make setup)
check: fmt vet test

## Run server locally
run:
	go run ./cmd/server

## Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html

## Show help
help:
	@echo "Available targets:"
	@echo "  setup        - Setup development environment (installs goimports, golangci-lint)"
	@echo "  build        - Build local binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  coverage     - Generate coverage report"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter (requires golangci-lint)"
	@echo "  vet          - Run go vet"
	@echo "  check        - Run all checks"
	@echo "  run          - Run server locally"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"
