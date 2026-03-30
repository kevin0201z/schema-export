.PHONY: build build-all clean test install

# Variables
BINARY_NAME=schema-export
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/schema-export

# Build for all platforms
build-all: build-linux build-windows build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/schema-export

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/schema-export

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/schema-export

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install binary to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/schema-export

# Run the application
run: build
	./$(BINARY_NAME)

# Development mode with hot reload (requires air)
dev:
	air

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Check for vulnerabilities
vuln:
	govulncheck ./...

# Generate test coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary for current platform"
	@echo "  build-all    - Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux  - Build for Linux AMD64"
	@echo "  build-windows- Build for Windows AMD64"
	@echo "  build-darwin - Build for macOS AMD64"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run tests"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  run          - Build and run the application"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  coverage     - Generate test coverage report"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help         - Show this help message"
