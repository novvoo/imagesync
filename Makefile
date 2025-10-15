# ImageSyncer Makefile
# 支持不同操作系统和架构的构建

# 变量定义
BINARY_NAME=imagesyncer
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标签
BUILD_TAGS=containers_image_openpgp

# 默认目标
.PHONY: all
all: build

# 构建目标
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -tags=$(BUILD_TAGS) -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" -o $(BINARY_NAME)$(EXT) .

# Windows 构建
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@$(MAKE) build EXT=.exe

# Linux 构建
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 $(MAKE) build

# macOS 构建
.PHONY: build-macos
build-macos:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 $(MAKE) build

# 多平台构建
.PHONY: build-all
build-all: build-windows build-linux build-macos
	@echo "All platforms built successfully!"

# 开发构建（包含调试信息）
.PHONY: build-dev
build-dev:
	@echo "Building development version..."
	@go build -tags=$(BUILD_TAGS) -race -ldflags "-X main.Version=$(VERSION)-dev -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" -o $(BINARY_NAME)$(EXT) .

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning build files..."
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	@go clean

# 运行程序
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)$(EXT)

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	@go test -tags=$(BUILD_TAGS) -v ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# 依赖管理
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# 安装依赖
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install -tags=$(BUILD_TAGS) .

# 显示帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-windows- Build for Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-macos  - Build for macOS"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-dev    - Build development version with race detection"
	@echo "  clean        - Clean build files"
	@echo "  run          - Build and run the program"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  install      - Install the binary"
	@echo "  help         - Show this help message"