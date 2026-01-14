@echo off
echo Starting NOFX Development Environment with Hot Reload...
echo.

REM Change to project directory
cd /d "e:\AI\nofx_Dev"

echo Building and starting development environment with hot reload...
docker-compose -f docker-compose.dev.watch.yml up -d --build

echo.
echo Checking running containers...
docker ps

echo.
echo NOFX Development Environment is now running!
echo - Backend API: http://localhost:8888
echo - Frontend UI: http://localhost:3300
echo.
echo Press any key to exit...
pause >nul