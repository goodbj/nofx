@echo off
echo ========================================
echo    停止 NOFX 开发版 Docker 服务
echo ========================================
echo.
echo 正在停止 NOFX 开发版服务...
echo.

docker-compose -f docker-compose.nofx-dev-chrome-hot.yml down

if %errorlevel% == 0 (
    echo.
    echo ========================================
    echo    NOFX 开发版服务已停止
    echo ========================================
    echo.
) else (
    echo.
    echo ========================================
    echo    停止服务时出现错误
    echo ========================================
    pause
)