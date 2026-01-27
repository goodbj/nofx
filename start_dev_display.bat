@echo off
setlocal

echo Starting nofx-dev-web (Development version with browser popup)...
echo Frontend: http://localhost:3300
echo Backend: http://localhost:8888
echo.

:: Check if Docker is running
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Docker is not running. Please start Docker Desktop first.
    pause
    exit /b 1
)

:: Navigate to the project directory
cd /d "%~dp0"

:: Start the services with docker-compose
echo Starting nofx-dev-frontend-display and nofx-dev-backend-display services...
docker-compose -f docker-compose.dev.display.yml up --build

pause