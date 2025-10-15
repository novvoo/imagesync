# ImageSyncer PowerShell 构建脚本
# 支持不同环境的构建

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# 变量定义
$BINARY_NAME = "imagesyncer"
$VERSION = "1.0.0"
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd_HH-mm-ss"
$BUILD_TAGS = "containers_image_openpgp"

function Show-Help {
    Write-Host "ImageSyncer PowerShell Build Script" -ForegroundColor Green
    Write-Host ""
    Write-Host "Available commands:" -ForegroundColor Yellow
    Write-Host "  build        - Build for Windows"
    Write-Host "  build-dev    - Build development version with race detection"
    Write-Host "  build-linux  - Build for Linux"
    Write-Host "  build-macos  - Build for macOS"
    Write-Host "  run          - Build and run the program"
    Write-Host "  test         - Run tests"
    Write-Host "  fmt          - Format code"
    Write-Host "  vet          - Run go vet"
    Write-Host "  deps         - Download and tidy dependencies"
    Write-Host "  clean        - Clean build files"
    Write-Host "  install      - Install the binary"
    Write-Host "  help         - Show this help message"
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [command]" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 build"
    Write-Host "  .\build.ps1 run"
    Write-Host "  .\build.ps1 build-linux"
}

function Build-Windows {
    Write-Host "Building $BINARY_NAME..." -ForegroundColor Green
    $result = go build -tags=$BUILD_TAGS -o "$BINARY_NAME.exe" .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful!" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function Build-Dev {
    Write-Host "Building development version..." -ForegroundColor Green
    $result = go build -tags=$BUILD_TAGS -race -o "$BINARY_NAME.exe" .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Development build successful!" -ForegroundColor Green
    } else {
        Write-Host "Development build failed!" -ForegroundColor Red
        exit 1
    }
}

function Build-Linux {
    Write-Host "Building for Linux..." -ForegroundColor Green
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $result = go build -tags=$BUILD_TAGS -o $BINARY_NAME .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Linux build successful!" -ForegroundColor Green
    } else {
        Write-Host "Linux build failed!" -ForegroundColor Red
        exit 1
    }
}

function Build-MacOS {
    Write-Host "Building for macOS..." -ForegroundColor Green
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    $result = go build -tags=$BUILD_TAGS -o $BINARY_NAME .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "macOS build successful!" -ForegroundColor Green
    } else {
        Write-Host "macOS build failed!" -ForegroundColor Red
        exit 1
    }
}

function Run-Program {
    Write-Host "Building and running $BINARY_NAME..." -ForegroundColor Green
    Build-Windows
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Running $BINARY_NAME..." -ForegroundColor Green
        & ".\$BINARY_NAME.exe"
    } else {
        Write-Host "Build failed, cannot run!" -ForegroundColor Red
        exit 1
    }
}

function Run-Tests {
    Write-Host "Running tests..." -ForegroundColor Green
    go test -tags=$BUILD_TAGS -v ./...
}

function Format-Code {
    Write-Host "Formatting code..." -ForegroundColor Green
    go fmt ./...
}

function Run-Vet {
    Write-Host "Running go vet..." -ForegroundColor Green
    go vet ./...
}

function Get-Dependencies {
    Write-Host "Downloading dependencies..." -ForegroundColor Green
    go mod download
    go mod tidy
}

function Clean-Build {
    Write-Host "Cleaning build files..." -ForegroundColor Green
    if (Test-Path "$BINARY_NAME.exe") { Remove-Item "$BINARY_NAME.exe" }
    if (Test-Path $BINARY_NAME) { Remove-Item $BINARY_NAME }
    go clean
}

function Install-Binary {
    Write-Host "Installing $BINARY_NAME..." -ForegroundColor Green
    Build-Windows
    if ($LASTEXITCODE -eq 0) {
        go install -tags=$BUILD_TAGS .
        Write-Host "Installation successful!" -ForegroundColor Green
    } else {
        Write-Host "Build failed, cannot install!" -ForegroundColor Red
        exit 1
    }
}

# 主逻辑
switch ($Command.ToLower()) {
    "build" { Build-Windows }
    "build-dev" { Build-Dev }
    "build-linux" { Build-Linux }
    "build-macos" { Build-MacOS }
    "run" { Run-Program }
    "test" { Run-Tests }
    "fmt" { Format-Code }
    "vet" { Run-Vet }
    "deps" { Get-Dependencies }
    "clean" { Clean-Build }
    "install" { Install-Binary }
    "help" { Show-Help }
    default { 
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Show-Help
        exit 1
    }
}
