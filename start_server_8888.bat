@echo off
chcp 65001 >nul
setlocal

echo 启动 NOFX 后端服务器 (端口 8888)...
echo ==============================

REM 设置环境变量
set SERVER_PORT=8888

echo 正在启动后端服务器，端口: %SERVER_PORT%
echo.

REM 导航到项目目录
cd /d "e:\AI\nofx"

REM 启动后端服务器
echo 启动命令: go run main.go
set OLLAMA_READ_TIMEOUT=270s && go run main.go

if errorlevel 1 (
    echo.
    echo 错误: 后端服务器启动失败
    pause
    exit /b 1
)

pause