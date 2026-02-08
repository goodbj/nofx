@echo off
REM 停止交易所代理服务脚本

echo 正在停止交易所代理服务...

cd /d "%~dp0\tools_docker_proxy"

docker compose down

if %ERRORLEVEL% EQU 0 (
    echo.
    echo 交易所代理服务已停止
    echo 容器名称: tools_docker_proxy
) else (
    echo.
    echo 错误: 停止服务时出现问题
    pause
    exit /b 1
)

pause