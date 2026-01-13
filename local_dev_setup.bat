@echo off
chcp 65001 >nul
echo.
echo ==============================================
echo        NOFX Local Development Setup
echo ==============================================
echo.

REM 检查Go是否已安装
echo Checking Go installation...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go 1.19+ from https://golang.org/dl/
    pause
    exit /b 1
)

REM 检查Node.js是否已安装
echo Checking Node.js installation...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Node.js is not installed or not in PATH
    echo Please install Node.js 18+ from https://nodejs.org/
    pause
    exit /b 1
)

REM 检查npm是否已安装
echo Checking npm installation...
npm --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: npm is not installed
    echo Please install npm with Node.js from https://nodejs.org/
    pause
    exit /b 1
)

REM 检查是否已有.env文件，如果没有则复制模板
if not exist ".env" (
    echo Creating .env file from template...
    if exist ".env.example" (
        copy .env.example .env
        echo Created .env from .env.example
    ) else (
        echo ERROR: Neither .env nor .env.example found
        pause
        exit /b 1
    )
)

REM 检查data目录是否存在
if not exist "data" (
    echo Creating data directory...
    mkdir data
)

echo.
echo Installing Go dependencies...
go mod tidy

echo.
echo Setting up frontend...
cd web
npm install
if %errorlevel% neq 0 (
    echo Error installing frontend dependencies
    pause
    exit /b 1
)
cd ..

echo.
echo Starting NOFX backend server in a new window...
start "NOFX Backend" cmd /k "go run main.go"

echo Waiting for backend to start...
ping -n 5 127.0.0.1 > nul

echo Starting NOFX frontend server in a new window...
cd web
start "NOFX Frontend" cmd /k "npm run dev"
cd ..

echo.
echo ==============================================
echo         Local Development Setup Complete!
echo ==============================================
echo.
echo  Backend:  http://localhost:8888
echo  Frontend: http://localhost:3300
echo  API Docs: http://localhost:8888/swagger/index.html
echo.
echo  To test your changes:
echo  1. Edit your Go files (e.g., in api/, trader/, decision/, etc.)
echo  2. Stop the backend (Ctrl+C in backend window)
echo  3. Run 'go run main.go' again to pick up changes
echo  4. Frontend updates are hot-reloaded automatically
echo.
echo ==============================================
echo.
pause