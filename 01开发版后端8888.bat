@echo off
chcp 65001 >nul
setlocal

set SERVER_PORT=8888

echo 启动 NOFX 后端服务器...
echo ============================

echo 正在启动后端服务器，端口: %SERVER_PORT%
echo.

cd /d "e:\AI\nofx_Dev"

echo 启动命令: go run main.go
set API_SERVER_PORT=%SERVER_PORT% && set JWT_SECRET=dev-jwt-secret-change-in-production && set DATA_ENCRYPTION_KEY=ZGV2LWRhdGEtZW5jcnlwdGlvbi1rZXktZGV2LWRhdGE= && go run main.go

if errorlevel 1 (
    echo.
    echo 错误: 后端服务器启动失败
    pause
    exit /b 1
)

pause