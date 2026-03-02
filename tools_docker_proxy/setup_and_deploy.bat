@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🚀 NOFX 透明代理 Docker 部署脚本
echo ========================================
echo.

REM 检查 Docker 是否已安装
echo [1/5] 检查 Docker 环境...
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

REM 检查 Docker Compose 是否已安装
echo [2/5] 检查 Docker Compose 环境...
docker-compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 错误: Docker Compose 未安装
    echo    请确保 Docker Desktop 的 Compose 功能已启用
    pause
    exit /b 1
)

echo ✅ Docker Compose 环境正常
echo.

REM 停止现有容器
echo [3/5] 停止现有容器...
docker-compose -f docker-compose.hot.yml down 2>nul
if %errorlevel% == 0 (
    echo ✅ 现有容器已停止
) else (
    echo ℹ️  未发现正在运行的容器
)
echo.

REM 设置环境变量
echo [4/5] 设置环境变量...
set PROXY_PORT=8081
echo ✅ 端口设置为: %PROXY_PORT%
echo.

REM 构建并启动服务
echo [5/5] 构建并启动透明代理服务（支持热更新）...
docker-compose -f docker-compose.hot.yml up -d --build

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo ✅ 部署成功!
    echo ========================================
    echo 服务信息:
    echo   - 容器名称: tools_docker_proxy
    echo   - 端口: %PROXY_PORT%
    echo   - 功能: 透明代理服务
    echo   - 热更新: 已启用
    echo   - 访问地址: http://localhost:%PROXY_PORT%
    echo.
    echo 健康检查: http://localhost:%PROXY_PORT%/health
    echo.
    echo 管理命令:
    echo   - 查看日志: docker logs -f tools_docker_proxy
    echo   - 停止服务: docker-compose -f docker-compose.hot.yml down
    echo   - 重启服务: docker restart tools_docker_proxy
    echo ========================================
) else (
    echo.
    echo ❌ 部署失败，请检查错误信息
    echo 可能的原因:
    echo   - 端口 %PROXY_PORT% 已被占用
    echo   - Docker 资源不足
    echo   - 网络连接问题
    docker logs tools_docker_proxy 2>nul
)

echo.
echo 按任意键退出...
pause >nul