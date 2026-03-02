@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🕵️  检查 NOFX 透明代理环境
echo ========================================
echo.

echo 检查 Docker 安装状态...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未安装
    echo    请先安装 Docker Desktop
    echo    下载地址: https://www.docker.com/products/docker-desktop
    echo.
    echo 当前无法部署透明代理服务到 Docker
    echo 但您可以使用以下方式运行服务：
    echo   - 本地运行: go run main.go
    echo   - 或先安装 Docker 后再部署
    goto end
) else (
    echo ✅ Docker 已安装
)

echo.
echo 检查 Docker 服务运行状态...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  Docker 服务未运行
    echo    请启动 Docker Desktop 应用程序
    echo.
    echo 当前无法部署透明代理服务到 Docker
    goto end
) else (
    echo ✅ Docker 服务正在运行
)

echo.
echo 检查 Docker Compose...
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  Docker Compose 未安装或不可用
    echo    请确保 Docker Desktop 的 Compose 功能已启用
    goto end
) else (
    echo ✅ Docker Compose 已安装
)

echo.
echo 检查透明代理容器...
docker ps --filter "name=tools_docker_proxy" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
if %errorlevel% equ 0 (
    set container_exists=1
    echo.
    echo 📊 透明代理服务状态: 已运行
) else (
    echo.
    echo 📊 透明代理服务状态: 未运行
)

echo.
echo ========================================
echo 环境摘要:
echo   - Docker 状态: 已安装并运行
echo   - 透明代理配置: 已准备就绪 (端口 8081)
echo   - 透明代理服务: %container_exists% 未运行
echo ========================================
echo.
echo 可选操作:
echo   1. 启动透明代理: .\QUICK_START.bat
echo   2. 检查主项目配置: 确认 .env 中 USE_BINANCE_PROXY=true
echo   3. 查看当前运行容器: docker ps -a
echo ========================================
:end
echo 按任意键退出...
pause >nul