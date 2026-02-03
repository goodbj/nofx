@echo off
chcp 65001 >nul
echo.
echo ============================================
echo        币安API代理服务重启脚本
echo ============================================
echo.

echo 正在停止币安代理服务...
docker stop binance-proxy-service >nul 2>&1
docker rm binance-proxy-service >nul 2>&1
echo [✓] 服务已停止

echo.
echo 正在重新启动币安代理服务...
call deploy_binance_proxy.bat