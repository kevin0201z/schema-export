# schema-export Makefile

# 变量定义
BINARY_NAME=schema-export
MAIN_PACKAGE=./cmd/schema-export
BUILD_DIR=build
VERSION?=dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 参数
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
GOFILES=$(wildcard *.go)
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# 并行构建支持
.NOTPARALLEL: clean
MAKEFLAGS += -j$(shell nproc 2>/dev/null || echo 4)

# 检查工具依赖
.PHONY: check-go
check-go:
	@echo "Checking Go installation..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Error: go command not found"; \
		exit 1; \
	fi
	@echo "Go version: $$(go version)"

# 检查 git 依赖
.PHONY: check-git
check-git:
	@echo "Checking Git installation..."
	@if ! command -v git >/dev/null 2>&1; then \
		echo "Warning: git command not found, commit will be 'unknown'"; \
	fi

# 检查所有依赖
.PHONY: check-deps
check-deps: check-go check-git

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build: check-deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR) || { echo "Failed to create build directory"; exit 1; }
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE) || { echo "Build failed"; exit 1; }
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 安装到 $GOPATH/bin
.PHONY: install
install: check-deps
	go install $(LDFLAGS) $(MAIN_PACKAGE) || { echo "Install failed"; exit 1; }

# 清理构建产物
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) || { echo "Failed to remove build directory"; exit 1; }
	@go clean || { echo "Go clean failed"; exit 1; }
	@echo "Clean complete"

# 运行测试
.PHONY: test
test: check-deps
	@echo "Running tests..."
	go test -v ./... || { echo "Tests failed"; exit 1; }

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage: check-deps
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./... || { echo "Tests failed"; exit 1; }
	go tool cover -html=coverage.out -o coverage.html || { echo "Coverage report generation failed"; exit 1; }
	@echo "Coverage report generated: coverage.html"

# 代码格式化
.PHONY: fmt
fmt: check-deps
	@echo "Formatting code..."
	go fmt ./... || { echo "Formatting failed"; exit 1; }

# 代码检查
.PHONY: vet
vet: check-deps
	@echo "Running go vet..."
	go vet ./... || { echo "Vet failed"; exit 1; }

# 静态检查 (需要安装 golangci-lint)
.PHONY: lint
lint: check-deps
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run || { echo "Linter failed"; exit 1; } \
	else \
		echo "golangci-lint not installed, skipping..."; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

# 下载依赖
.PHONY: deps
deps: check-deps
	@echo "Downloading dependencies..."
	go mod download || { echo "Dependency download failed"; exit 1; }
	go mod tidy || { echo "Dependency tidy failed"; exit 1; }

# 跨平台编译
.PHONY: build-all
build-all: check-deps build-linux build-windows build-darwin

.PHONY: build-linux
build-linux: check-deps
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR) || { echo "Failed to create build directory"; exit 1; }
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE) || { echo "Linux build failed"; exit 1; }

.PHONY: build-windows
build-windows: check-deps
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR) || { echo "Failed to create build directory"; exit 1; }
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE) || { echo "Windows build failed"; exit 1; }

.PHONY: build-darwin
build-darwin: check-deps
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR) || { echo "Failed to create build directory"; exit 1; }
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE) || { echo "macOS amd64 build failed"; exit 1; }
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE) || { echo "macOS arm64 build failed"; exit 1; }

# 运行示例
.PHONY: run
run: build
	$(BUILD_DIR)/$(BINARY_NAME) --help

# 开发模式运行
.PHONY: dev
dev: check-deps
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
