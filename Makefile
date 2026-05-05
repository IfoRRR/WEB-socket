.PHONY: help build run clean test dev

help:
	@echo "Available commands:"
	@echo "  make build    - Build the server binary"
	@echo "  make run      - Run the server"
	@echo "  make dev      - Run in development mode (with hot reload)"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter"

build:
	@echo "Building server..."
	go build -o server ./cmd/server
	@echo "Build complete: ./server"

run: build
	@echo "Running server..."
	./server

dev:
	@echo "Running in development mode..."
	go run ./cmd/server

clean:
	@echo "Cleaning..."
	rm -f server
	rm -f *.db
	@echo "Clean complete"

test:
	@echo "Running tests..."
	go test -v ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	go vet ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
