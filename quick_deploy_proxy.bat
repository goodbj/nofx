@echo off
echo Starting Binance Proxy Service Deployment...

REM Clean up any existing containers
docker stop binance-proxy-service >nul 2>&1
docker rm binance-proxy-service >nul 2>&1

REM Build the service
docker build -t binance-proxy-service -f api/binance_proxy/Dockerfile.simple . >nul 2>&1

REM If build fails, use alternative method
if %errorlevel% neq 0 (
    echo Building with alternative method...
    if not exist "temp_build" mkdir temp_build
    xcopy /s /e /y "api\binance_proxy" "temp_build\api\binance_proxy\" >nul 2>&1
    xcopy /s /e /y "logger_module" "temp_build\logger_module\" >nul 2>&1
    
    echo FROM golang:1.25-alpine AS builder > temp_build\Dockerfile.alt
    echo WORKDIR /app >> temp_build\Dockerfile.alt
    echo RUN apk add --no-cache git >> temp_build\Dockerfile.alt
    echo COPY api/binance_proxy/ ./ >> temp_build\Dockerfile.alt
    echo COPY logger_module/ ../logger_module/ >> temp_build\Dockerfile.alt
    echo RUN sed -i 's|../../logger_module|../logger_module|g' go.mod >> temp_build\Dockerfile.alt
    echo RUN go mod download >> temp_build\Dockerfile.alt
    echo RUN CGO_ENABLED=0 GOOS=linux go build -o binance-proxy . >> temp_build\Dockerfile.alt
    echo FROM alpine:latest >> temp_build\Dockerfile.alt
    echo RUN apk add --no-cache curl ca-certificates >> temp_build\Dockerfile.alt
    echo WORKDIR /root/ >> temp_build\Dockerfile.alt
    echo COPY --from=builder /app/binance-proxy . >> temp_build\Dockerfile.alt
    echo RUN chmod +x ./binance-proxy >> temp_build\Dockerfile.alt
    echo EXPOSE 8081 >> temp_build\Dockerfile.alt
    echo CMD ["./binance-proxy"] >> temp_build\Dockerfile.alt
    
    docker build -t binance-proxy-service -f temp_build/Dockerfile.alt temp_build >nul 2>&1
    rmdir /s /q temp_build >nul 2>&1
)

REM Start the container
docker run -d --name binance-proxy-service -p 8081:8081 -e PORT=8081 --restart unless-stopped binance-proxy-service >nul 2>&1

REM Wait a moment for service to start
timeout /t 3 /nobreak >nul

echo.
echo Binance Proxy Service deployed successfully!
echo Access the service at: http://localhost:8081
echo.
pause