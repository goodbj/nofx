@echo off

chcp 65001 >nul

setlocal enabledelayedexpansion

set SERVER_PORT=8888

echo 启动 NOFX 后端服务器...
echo ============================

echo 正在启动后端服务器，端口: %SERVER_PORT%
echo.

cd /d "e:\AI\nofx_Dev"

echo 启动命令: go run main.go
echo 注意: 将使用 .env 文件中的配置，包括 DATA_ENCRYPTION_KEY 和 JWT_SECRET
set API_SERVER_PORT=%SERVER_PORT% && go run -mod=mod main.go

if errorlevel 1 (
    echo.
    echo 错误: 后端服务器启动失败
    pause
    exit /b 1
)

pause