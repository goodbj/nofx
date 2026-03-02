@echo off
chcp 65001 >nul
echo.
echo ========================================
echo ⚡ 快速启动 NOFX 透明代理服务
echo ========================================
echo.

REM 检查 Docker
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未安装或未运行
    echo    请先安装 Docker Desktop
    pause
    exit /b 1
)

REM 停止现有服务
docker-compose -f docker-compose.hot.yml down 2>nul

REM 启动服务
echo 启动透明代理服务...
docker-compose -f docker-compose.hot.yml up -d --build

if %errorlevel% equ 0 (
    echo.
    echo ✅ 透明代理服务已启动
    echo 🌐 地址: http://localhost:8081
    echo 🏷️  端口: 8081 (支持热更新)
    echo 🔄 正在检查健康状态...
    timeout /t 5 /nobreak >nul
    docker logs tools_docker_proxy --tail 5
) else (
    echo ❌ 启动失败
    docker logs tools_docker_proxy 2>nul
)

echo.
echo ========================================
echo 快速命令:
echo   - 查看日志: docker logs -f tools_docker_proxy
echo   - 停止服务: docker-compose -f docker-compose.hot.yml down
echo ========================================
echo.
pause