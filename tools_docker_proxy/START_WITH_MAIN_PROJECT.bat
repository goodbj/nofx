@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🚀 启动 NOFX 透明代理及主项目
echo ========================================
echo.

REM 检查 Docker 是否已安装
echo [1/4] 检查 Docker 环境...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: Docker 未安装或未运行
    echo    请先安装 Docker Desktop 并启动服务
    echo    下载地址: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo ✅ Docker 环境正常
echo.

REM 停止现有容器
echo [2/4] 停止现有透明代理容器...
docker-compose -f docker-compose.hot.yml down 2>nul
if %errorlevel% == 0 (
    echo ✅ 现有容器已停止
) else (
    echo ℹ️  未发现正在运行的透明代理容器
)
echo.

REM 启动透明代理服务
echo [3/4] 启动透明代理服务（支持热更新）...
docker-compose -f docker-compose.hot.yml up -d --build

if %errorlevel% neq 0 (
    echo ❌ 透明代理服务启动失败
    docker logs tools_docker_proxy 2>nul
    pause
    exit /b 1
)

echo ✅ 透明代理服务启动成功
echo.

REM 等待代理服务完全启动
echo [4/4] 等待服务启动完成...
timeout /t 10 /nobreak >nul

REM 检查代理服务健康状态
echo 测试透明代理健康检查...
curl -s http://localhost:8081/health
if %errorlevel% equ 0 (
    echo ✅ 透明代理服务运行正常
) else (
    echo ⚠️  透明代理服务可能未正常运行
)

echo.
echo ========================================
echo 🎉 部署完成!
echo ========================================
echo 透明代理服务信息:
echo   - 容器名称: tools_docker_proxy
echo   - 端口: 8081
echo   - 功能: 透明代理服务
echo   - 热更新: 已启用
echo   - 访问地址: http://localhost:8081
echo.
echo 健康检查: http://localhost:8081/health
echo.
echo 接下来您可以：
echo   1. 在新窗口中启动主项目: go run main.go
echo   2. 或运行主项目的其他启动脚本
echo   3. 确保 .env 文件中的 USE_BINANCE_PROXY=true
echo      以及 BINANCE_PROXY_URL=http://localhost:8081
echo ========================================
echo.
echo 按任意键启动完成确认...
pause >nul