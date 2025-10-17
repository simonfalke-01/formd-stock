.PHONY: build run test clean fmt lint install docker

# Build variables
BINARY_NAME=formd-stock
VERSION?=1.0.0
BUILD_FLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Build the application
build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .

# Run the application
run:
	go run .

# Run with config file
run-config:
	go run . -config config.json

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Lint code
lint:
	golangci-lint run

# Install dependencies
install:
	go mod download
	go mod tidy

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Docker build
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run --rm --env-file .env $(BINARY_NAME):$(VERSION)

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  run-config  - Run with config.json"
	@echo "  test        - Run tests with coverage"
	@echo "  clean       - Remove build artifacts"
	@echo "  fmt         - Format code"
	@echo "  lint        - Lint code"
	@echo "  install     - Install dependencies"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  docker      - Build Docker image"
	@echo "  docker-run  - Run Docker container"
