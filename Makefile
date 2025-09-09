# Makefile for Autosphere MCP Golang Server

.PHONY: build clean test run run-http deps fmt vet lint

# Variables
BINARY_NAME=autosphere-mcp-server
BUILD_DIR=./cmd/server
PORT=8080

# Default target
all: deps fmt vet build

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "🔍 Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "🧹 Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, skipping..."; \
	fi

# Build the binary
build: deps
	@echo "🔨 Building binary..."
	go build -o $(BINARY_NAME) $(BUILD_DIR)
	@echo "✅ Build successful! Binary: ./$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean

# Run with STDIO transport
run: build
	@echo "🚀 Running server with STDIO transport..."
	./$(BINARY_NAME)

# Run with HTTP transport
run-http: build
	@echo "🌐 Running server with HTTP transport on port $(PORT)..."
	./$(BINARY_NAME) -http localhost:$(PORT)

# Test build without running
test-build:
	@echo "🧪 Testing build..."
	go build -o /tmp/$(BINARY_NAME) $(BUILD_DIR)
	@echo "✅ Build test successful!"
	rm -f /tmp/$(BINARY_NAME)

# Run tests (when tests are added)
test:
	@echo "🧪 Running tests..."
	go test ./...

# Development setup
dev-setup: deps
	@echo "🛠️  Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "📥 Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "✅ Development environment ready!"

# Show help
help:
	@echo "Autosphere MCP Golang Server - Available Commands:"
	@echo ""
	@echo "  make build      - Build the binary"
	@echo "  make run        - Run with STDIO transport"
	@echo "  make run-http   - Run with HTTP transport (port 8080)"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make fmt        - Format code"
	@echo "  make vet        - Vet code"
	@echo "  make lint       - Lint code (requires golangci-lint)"
	@echo "  make deps       - Download dependencies"
	@echo "  make dev-setup  - Setup development environment"
	@echo "  make help       - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make run-http PORT=3000  - Run HTTP server on port 3000"
