@echo off
chcp 65001 >nul
echo.
echo ============================================
echo        币安API代理服务快速部署脚本
echo ============================================
echo.

REM 检查Docker是否已安装
echo [1/5] 检查Docker环境...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker未安装或未添加到PATH环境变量
    echo    请先安装Docker Desktop并重启系统
    pause
    exit /b 1
)

echo [OK] Docker环境检查通过

REM 检查Docker服务是否运行
echo [2/5] 检查Docker服务状态...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker服务未运行
    echo    请启动Docker Desktop并等待服务就绪
    pause
    exit /b 1
)

echo [OK] Docker服务运行正常

REM 进入项目目录
echo [3/5] 进入项目目录...
cd /d "%~dp0"
if %errorlevel% neq 0 (
    echo [ERROR] 无法进入项目目录
    pause
    exit /b 1
)

REM 检查并停止已存在的容器
echo [4/5] 检查并清理已存在的容器...
docker stop binance-proxy-service >nul 2>&1
docker rm binance-proxy-service >nul 2>&1
echo [OK] 清理完成

REM 构建并启动币安代理服务
echo [5/5] 构建并启动币安代理服务...
docker build -t binance-proxy-service -f api/binance_proxy/Dockerfile.simple . >nul 2>&1

REM 如果上面的构建失败，使用备用方案
if %errorlevel% neq 0 (
    echo [INFO] Dockerfile.simple构建失败，尝试使用项目根目录构建方式...

    REM 创建临时构建上下文
    if not exist "build_context" mkdir build_context
    xcopy /s /e /y "api\binance_proxy" "build_context\api\binance_proxy\" >nul 2>&1
    xcopy /s /e /y "logger_module" "build_context\logger_module\" >nul 2>&1

    REM 创建专用Dockerfile
    echo FROM golang:1.25-alpine AS builder > build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo WORKDIR /app >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN apk add --no-cache git >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo COPY api/binance_proxy/ ./ >> build_context\Dockerfile.proxy
    echo COPY logger_module/ ../logger_module/ >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN sed -i 's|../../logger_module|../logger_module|g' go.mod >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN go mod download >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN CGO_ENABLED=0 GOOS=linux go build -o binance-proxy . >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo FROM alpine:latest >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN apk add --no-cache curl ca-certificates >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo WORKDIR /root/ >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo COPY --from=builder /app/binance-proxy . >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo RUN chmod +x ./binance-proxy >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo EXPOSE 8082 >> build_context\Dockerfile.proxy
    echo. >> build_context\Dockerfile.proxy
    echo CMD ["./binance-proxy"] >> build_context\Dockerfile.proxy

    REM 构建服务
    docker build -t binance-proxy-service -f build_context/Dockerfile.proxy build_context >nul 2>&1
    if %errorlevel% neq 0 (
        echo [ERROR] 构建失败
        rmdir /s /q build_context >nul 2>&1
        pause
        exit /b 1
    )
    rmdir /s /q build_context >nul 2>&1
)

REM 启动容器
docker run -d ^
    --name binance-proxy-service ^
    -p 8081:8082 ^
    -e PORT=8082 ^
    -e GIN_MODE=release ^
    -e DEBUG=false ^
    --restart unless-stopped ^
    binance-proxy-service >nul 2>&1

if %errorlevel% neq 0 (
    echo [ERROR] 启动容器失败
    pause
    exit /b 1
)

REM 等待服务启动
timeout /t 5 /nobreak >nul

REM 检查容器状态
for /f %%i in ('docker inspect -f "{{.State.Running}}" binance-proxy-service 2^>nul') do set running=%%i
if "%running%" neq "true" (
    echo [ERROR] 容器启动失败
    docker logs binance-proxy-service
    pause
    exit /b 1
)

REM 测试服务
echo.
echo [SUCCESS] 服务部署完成！
echo.
echo 服务信息：
echo   容器名称: binance-proxy-service
echo   监听端口: 8081 (映射到容器内8082端口)
echo   功能: 处理币安API请求，绕过地区限制
echo.
echo 测试连接: 
ping -n 1 127.0.0.1 >nul
curl -s http://localhost:8081/health >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 服务连接测试成功！
) else (
    echo [WARN] 服务可能尚未完全启动，请稍等片刻后重试
)

echo.
echo ============================================
echo 部署完成！币安代理服务已启动并监听端口8081
echo ============================================
echo.
pause