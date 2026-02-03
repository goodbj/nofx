@echo off
echo.
echo ============================================
echo    NOFX 代理服务停止脚本
echo ============================================
echo.

echo 正在停止代理服务...
echo.

 REM 停止代理服务
docker-compose -f ./docker-compose.proxy.yml down

if %errorlevel% equ 0 (
    echo.
    echo 代理服务已停止！
) else (
    echo.
    echo 停止服务时出现错误
    docker-compose -f ./docker-compose.proxy.yml ps
)

echo.
echo 按任意键退出...
pause >nul