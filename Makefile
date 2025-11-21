.PHONY: build clean install test run help

VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.GitCommit=$(GIT_COMMIT)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	@echo "Building ssh-notes-server..."
	@go build -ldflags "$(LDFLAGS)" -o ssh-notes-server
	@echo "✓ Build complete"

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ssh-notes-server-linux
	@echo "✓ Linux build complete"

build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ssh-notes-server-darwin
	@echo "✓ macOS build complete"

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ssh-notes-server-windows.exe
	@echo "✓ Windows build complete"

build-all: build-linux build-darwin build-windows ## Build for all platforms

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f ssh-notes-server ssh-notes-server-*
	@echo "✓ Clean complete"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

install: build ## Install to /usr/local/bin
	@echo "Installing..."
	@sudo cp ssh-notes-server /usr/local/bin/
	@sudo cp ssh-notes /usr/local/bin/ 2>/dev/null || true
	@echo "✓ Installation complete"

run: build ## Build and run the server
	@./ssh-notes-server -port 2222 -data ./data

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"

lint: fmt vet ## Run linters

release: clean build-all ## Create release builds
	@echo "Creating release..."
	@mkdir -p release
	@cp ssh-notes-server-linux release/
	@cp ssh-notes-server-darwin release/
	@cp ssh-notes-server-windows.exe release/
	@cp README.md LICENSE release/
	@echo "✓ Release created in release/"

