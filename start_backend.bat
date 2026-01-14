@echo off
chcp 65001 >nul
setlocal

echo 启动 NOFX 后端服务器...
echo ==============================

REM 检查是否已设置环境变量，如果没有则使用默认值
if "%SERVER_PORT%"=="" set SERVER_PORT=8888

echo 正在启动后端服务器，端口: %SERVER_PORT%
echo.

REM 导航到项目目录
cd /d "e:\AI\nofx"

REM 启动后端服务器
REM echo 启动命令: go run main.go
REM go run main.go

cd e:\AI\nofx
$env:CGO_ENABLED="1"
go run main.go

if errorlevel 1 (
    echo.
    echo 错误: 后端服务器启动失败
    pause
    exit /b 1
)

pause