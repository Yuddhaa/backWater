# Default command
all: help

# Run the app in development mode with hot-reload
dev: #install-air .air.toml
	@echo "Starting server with hot-reload (Air)..."
	@air

# Build the Go binary locally
build:
	@echo "Building Go binary..."
	# @mkdir -p tmp
	@go build -o ./tmp/server ./...

run: build
	./tmp/server

# Install the 'air' hot-reload tool
install-air:
	@echo "Installing/Updating 'air'..."
	@go install github.com/air-verse/air@latest

# Show help menu
help:
	@echo "Available commands:"
	@echo "  make dev          - Start local dev server with hot-reload"
	@echo "  make build        - Build the Go binary locally"
	@echo "  make run          - Build and Run the Go binary locally"
	@echo "  make install-air  - Install the 'air' tool"
