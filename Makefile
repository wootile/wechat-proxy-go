# WeChat API Proxy Server Makefile

.PHONY: help clean build build-linux build-windows build-all test run

# Default target
.DEFAULT_GOAL := help

# Variable definitions
BINARY_NAME := wechat-proxy
BUILD_DIR := build

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X main.Version=$(VERSION) -w -s

# Help information
help: ## Show help information
	@echo "WeChat API Proxy Server - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "✅ Clean completed"

# Local build
build: ## Build local version
	@echo "🔨 Building local version..."
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) main.go
	@echo "✅ Build completed: $(BINARY_NAME)"

# Linux build
build-linux: ## Build Linux version
	@echo "🔨 Building Linux version..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "✅ Linux build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Windows build
build-windows: ## Build Windows version
	@echo "🔨 Building Windows version..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✅ Windows build completed: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# Build all platforms
build-all: build-linux build-windows ## Build all platform versions
	@echo "🎉 All platform builds completed"

# Run tests
test: build ## Build and test connection
	@echo "🧪 Testing proxy server..."
	@echo "Starting proxy server (background)..."
	@./$(BINARY_NAME) &
	@sleep 2
	@echo "Testing proxy connection..."
	@curl -x http://localhost:8080 --connect-timeout 5 -s -o /dev/null -w "Status code: %{http_code}\n" https://api.weixin.qq.com || echo "Connection test failed"
	@echo "Stopping proxy server..."
	@pkill -f $(BINARY_NAME) || true
	@echo "✅ Test completed"

# Local run
run: build ## Build and run
	@echo "🚀 Starting proxy server..."
	@./$(BINARY_NAME) 