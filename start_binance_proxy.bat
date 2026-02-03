@echo off
chcp 65001 >nul

echo 快速启动币安代理服务...

:: 检查容器是否存在
docker ps -a | findstr binance-proxy-container >nul
if %errorLevel% neq 0 (
    echo ❌ 未找到代理服务容器，正在运行安装脚本...
    call install_binance_proxy.bat
    exit /b
)

:: 检查服务是否运行
docker ps | findstr binance-proxy-container >nul
if %errorLevel% equ 0 (
    echo ✅ 服务已在运行
    echo 服务地址: http://localhost:8081/health
) else (
    echo 启动服务...
    docker start binance-proxy-container >nul 2>&1
    if %errorLevel% equ 0 (
        echo ✅ 服务启动成功
        timeout /t 3 /nobreak >nul
        curl -s http://localhost:8081/health >nul 2>&1
        if %errorLevel% equ 0 (
            echo ✅ 服务连接正常
            echo 服务地址: http://localhost:8081/health
        ) else (
            echo ⚠️  服务启动中，请稍后检查
        )
    ) else (
        echo ❌ 启动失败
    )
)

echo.
echo 按任意键退出...
pause >nul