@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版 Docker 服务停止脚本
echo ===============================================
echo.
echo 正在停止服务...
echo.

:: 停止服务
docker compose -f docker-compose.nofx-dev.yml down

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    NOFX 开发版服务已成功停止！
    echo ===============================================
    echo.
) else (
    echo.
    echo 错误：停止服务时出现问题
    echo 请检查错误信息
)

pause