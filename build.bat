@echo off
REM ImageSyncer Windows 构建脚本
REM 支持不同环境的构建

setlocal

REM 变量定义
set BINARY_NAME=imagesyncer
set VERSION=1.0.0
set BUILD_TIME=%date:~0,4%-%date:~5,2%-%date:~8,2%_%time:~0,2%-%time:~3,2%-%time:~6,2%
set BUILD_TAGS=containers_image_openpgp

REM 检查参数
if "%1"=="" goto :help
if "%1"=="help" goto :help
if "%1"=="build" goto :build
if "%1"=="build-dev" goto :build-dev
if "%1"=="build-linux" goto :build-linux
if "%1"=="build-macos" goto :build-macos
if "%1"=="run" goto :run
if "%1"=="test" goto :test
if "%1"=="fmt" goto :fmt
if "%1"=="vet" goto :vet
if "%1"=="deps" goto :deps
if "%1"=="clean" goto :clean
if "%1"=="install" goto :install
goto :help

:build
echo Building %BINARY_NAME%...
go build -tags=%BUILD_TAGS% -o %BINARY_NAME%.exe .
if %ERRORLEVEL% EQU 0 (
    echo Build successful!
) else (
    echo Build failed!
    exit /b 1
)
goto :end

:build-dev
echo Building development version...
go build -tags=%BUILD_TAGS% -race -o %BINARY_NAME%.exe .
if %ERRORLEVEL% EQU 0 (
    echo Development build successful!
) else (
    echo Development build failed!
    exit /b 1
)
goto :end

:build-linux
echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -tags=%BUILD_TAGS% -o %BINARY_NAME% .
if %ERRORLEVEL% EQU 0 (
    echo Linux build successful!
) else (
    echo Linux build failed!
    exit /b 1
)
goto :end

:build-macos
echo Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -tags=%BUILD_TAGS% -o %BINARY_NAME% .
if %ERRORLEVEL% EQU 0 (
    echo macOS build successful!
) else (
    echo macOS build failed!
    exit /b 1
)
goto :end

:run
echo Building and running %BINARY_NAME%...
call :build
if %ERRORLEVEL% EQU 0 (
    echo Running %BINARY_NAME%...
    %BINARY_NAME%.exe
) else (
    echo Build failed, cannot run!
    exit /b 1
)
goto :end

:test
echo Running tests...
go test -tags=%BUILD_TAGS% -v ./...
goto :end

:fmt
echo Formatting code...
go fmt ./...
goto :end

:vet
echo Running go vet...
go vet ./...
goto :end

:deps
echo Downloading dependencies...
go mod download
go mod tidy
goto :end

:clean
echo Cleaning build files...
if exist %BINARY_NAME%.exe del %BINARY_NAME%.exe
if exist %BINARY_NAME% del %BINARY_NAME%
go clean
goto :end

:install
echo Installing %BINARY_NAME%...
call :build
if %ERRORLEVEL% EQU 0 (
    go install -tags=%BUILD_TAGS% .
    echo Installation successful!
) else (
    echo Build failed, cannot install!
    exit /b 1
)
goto :end

:help
echo ImageSyncer Windows Build Script
echo.
echo Available commands:
echo   build        - Build for Windows
echo   build-dev    - Build development version with race detection
echo   build-linux  - Build for Linux
echo   build-macos  - Build for macOS
echo   run          - Build and run the program
echo   test         - Run tests
echo   fmt          - Format code
echo   vet          - Run go vet
echo   deps         - Download and tidy dependencies
echo   clean        - Clean build files
echo   install      - Install the binary
echo   help         - Show this help message
echo.
echo Usage: build.bat [command]
echo.
echo Examples:
echo   build.bat build
echo   build.bat run
echo   build.bat build-linux
goto :end

:end
endlocal
