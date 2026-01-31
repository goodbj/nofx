@echo off
echo Starting NOFX System with Binance Proxy...

echo.
echo Step 1: Starting Binance Proxy Service in Docker...
docker start binance-proxy
if %errorlevel% neq 0 (
    echo Error starting binance-proxy container, attempting to run it...
    docker run -d --name binance-proxy -p 8081:8082 binance-proxy
)

timeout /t 5 /nobreak >nul

echo.
echo Step 2: Starting Backend Service on port 8888...
start cmd /k "cd /d %~dp0 && go run main.go"

echo.
echo All services started!
echo - Binance Proxy: http://localhost:8081/health
echo - Backend: http://localhost:8888/api/health
echo.
echo To start frontend, run: cd web && npm run dev
pause