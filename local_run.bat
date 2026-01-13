@echo off
chcp 65001 >nul
echo.
echo ==============================================
echo      NOFX Local Development Runner
echo ==============================================
echo.

REM 确保 data 目录存在
if not exist "data" (
    echo Creating data directory...
    mkdir data
)

REM 检查 .env 文件
if not exist ".env" (
    if exist ".env.example" (
        echo Creating .env from template...
        copy .env.example .env
    )
)

echo Starting NOFX backend server...
echo Press Ctrl+C to stop the server
echo.

REM 运行后端服务
go run main.go