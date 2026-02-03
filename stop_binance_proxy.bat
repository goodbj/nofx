@echo off
chcp 65001 >nul
echo.
echo ============================================
echo        币安API代理服务停止脚本
echo ============================================
echo.

echo 正在停止币安代理服务...
docker stop binance-proxy-service >nul 2>&1
docker rm binance-proxy-service >nul 2>&1

if %errorlevel% equ 0 (
    echo [✓] 币安代理服务已停止
) else (
    echo [✓] 服务可能未运行或已停止
)

echo.
echo ============================================
echo 服务已停止
echo ============================================
echo.
pause