@echo off
echo Starting NOFX System...

echo.
echo Starting Backend Service on port 8888...
start cmd /k "cd /d %~dp0 && go run main.go"

echo.
echo All services started!
echo - Backend: http://localhost:8888/api/health
echo.
echo To start frontend, run: cd web && npm run dev
pause