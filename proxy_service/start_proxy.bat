@echo off
echo.
echo ============================================
echo    NOFX 代理服务启动脚本
echo ============================================
echo.

echo 正在启动代理服务...
echo.

 REM 检查是否安装了 Docker
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未检测到 Docker，请先安装 Docker
    pause
    exit /b 1
)

 REM 启动代理服务
docker-compose -f ./docker-compose.proxy.yml up -d

if %errorlevel% equ 0 (
    echo.
    echo 代理服务启动成功！
    echo 服务地址: http://localhost:8082
    echo 健康检查: http://localhost:8082/health
    echo.
    echo 正在检查服务状态...
    timeout /t 5 /nobreak >nul
    docker-compose -f ./docker-compose.proxy.yml ps
) else (
    echo.
    echo 启动失败，请检查错误信息
)

echo.
echo 按任意键退出...
pause >nul