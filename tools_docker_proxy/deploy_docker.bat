@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo 🚀 透明代理服务 Docker 部署脚本
echo ========================================
echo.

REM 检查Docker是否运行
echo 🔍 检查 Docker 环境...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未运行或未安装，请先启动 Docker Desktop
    pause
    exit /b 1
)
echo ✅ Docker 环境正常
echo.

REM 停止现有容器
echo 🛑 停止现有代理容器...
docker stop tools_docker_proxy >nul 2>&1
docker rm tools_docker_proxy >nul 2>&1
echo ✅ 旧容器已清理
echo.

REM 构建镜像
echo 🏗️  构建 Docker 镜像...
docker build -t nofx-proxy:latest .
if %errorlevel% neq 0 (
    echo ❌ 镜像构建失败
    pause
    exit /b 1
)
echo ✅ 镜像构建成功
echo.

REM 启动容器
echo 🚀 启动代理容器...
docker run -d ^
  --name tools_docker_proxy ^
  --restart unless-stopped ^
  -p 8081:8081 ^
  -e PORT=8081 ^
  -e GIN_MODE=release ^
  --network nofx-proxy-network ^
  nofx-proxy:latest

if %errorlevel% neq 0 (
    echo ❌ 容器启动失败
    pause
    exit /b 1
)

echo ✅ 容器启动成功
echo.

REM 等待服务启动
echo ⏱️  等待服务启动...
timeout /t 5 /nobreak >nul

REM 检查服务状态
echo 🔍 检查服务状态...
docker logs tools_docker_proxy --tail 10

echo.
echo ========================================
echo 🎉 部署完成！
echo ========================================
echo 容器名称: tools_docker_proxy
echo 访问端口: 8081
echo 健康检查: curl http://localhost:8081/health
echo 查看日志: docker logs tools_docker_proxy
echo 停止服务: docker stop tools_docker_proxy
echo ========================================

pause