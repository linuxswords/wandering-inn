.PHONY: build run clean install help fmt lint

# Binary name
BINARY_NAME=wandering-inn-epub

# Default target
all: build

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/wandering-inn

# Run the application
run:
	go run ./cmd/wandering-inn

# Install dependencies
install:
	go mod tidy

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f *.epub

# Format code
fmt:
	gofmt -s -w .

# Run linter
lint:
	go vet ./...

# Run tests (if any exist)
test:
	go test ./...

# Build and run
run-binary: build
	./$(BINARY_NAME)

# Help target
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application directly with 'go run'"
	@echo "  run-binary  - Build and run the binary"
	@echo "  install     - Install/update dependencies"
	@echo "  clean       - Clean build artifacts and generated files"
	@echo "  fmt         - Format Go code"
	@echo "  lint        - Run linter (go vet)"
	@echo "  test        - Run tests"
	@echo "  help        - Show this help message"
