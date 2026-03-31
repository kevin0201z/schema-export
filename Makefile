# schema-export Makefile

# 变量定义
BINARY_NAME=schema-export
MAIN_PACKAGE=./cmd/schema-export
BUILD_DIR=./build
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 参数
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/build
GOFILES=$(wildcard *.go)
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 安装到 $GOPATH/bin
.PHONY: install
install:
	go install $(LDFLAGS) $(MAIN_PACKAGE)

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# 静态检查 (需要安装 golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"; \
	fi

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 跨平台编译
.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

# 运行示例
.PHONY: run
run: build
	$(BUILD_DIR)/$(BINARY_NAME) --help

# 开发模式运行
.PHONY: dev
dev:
	go run $(MAIN_PACKAGE) export --help

# 查看版本
.PHONY: version
version: build
	$(BUILD_DIR)/$(BINARY_NAME) version

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make install      - Install to \$$GOPATH/bin"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage- Run tests with coverage report"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run linter (golangci-lint)"
	@echo "  make deps         - Download and tidy dependencies"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make build-linux  - Build for Linux"
	@echo "  make build-windows- Build for Windows"
	@echo "  make build-darwin - Build for macOS"
	@echo "  make run          - Build and run"
	@echo "  make dev          - Run in development mode"
	@echo "  make version      - Show version"
	@echo "  make help         - Show this help"
