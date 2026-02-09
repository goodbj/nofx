@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:menu
cls
echo ========================================
echo 🐳 透明代理 Docker 管理菜单
echo ========================================
echo 1. 🔧 部署生产环境 (无热更新)
echo 2. 🔄 部署开发环境 (支持热更新)
echo 3. 📊 查看容器状态
echo 4. 📜 查看实时日志
echo 5. 🛑 停止容器
echo 6. 🔁 重启容器
echo 7. 🧹 清理所有相关资源
echo 8. 🚀 快速启动健康检查
echo 0. 退出
echo ========================================
set /p choice=请选择操作 (0-8): 

if "%choice%"=="1" goto deploy_prod
if "%choice%"=="2" goto deploy_dev
if "%choice%"=="3" goto status
if "%choice%"=="4" goto logs
if "%choice%"=="5" goto stop
if "%choice%"=="6" goto restart
if "%choice%"=="7" goto cleanup
if "%choice%"=="8" goto health_check
if "%choice%"=="0" goto :eof
goto menu

:deploy_prod
echo.
echo 🚀 部署生产环境...
call deploy_docker.bat
pause
goto menu

:deploy_dev
echo.
echo 🔄 部署开发环境...
echo 停止现有容器...
docker-compose -f docker-compose.dev.yml down
echo 构建并启动开发环境...
docker-compose -f docker-compose.dev.yml up -d --build
if %errorlevel% equ 0 (
    echo ✅ 开发环境部署成功
    echo 容器将自动监听代码变化并重启
) else (
    echo ❌ 部署失败
)
pause
goto menu

:status
echo.
echo 📊 容器状态:
docker ps -a --filter "name=tools_docker_proxy"
echo.
echo 🌐 网络状态:
docker network ls --filter "name=nofx-proxy-network"
pause
goto menu

:logs
echo.
echo 📜 实时日志 (按 Ctrl+C 退出):
docker logs -f tools_docker_proxy
goto menu

:stop
echo.
echo 🛑 停止容器...
docker stop tools_docker_proxy
echo ✅ 容器已停止
pause
goto menu

:restart
echo.
echo 🔁 重启容器...
docker restart tools_docker_proxy
echo ✅ 容器已重启
timeout /t 3 /nobreak >nul
docker logs tools_docker_proxy --tail 5
pause
goto menu

:cleanup
echo.
echo 🧹 清理资源...
docker stop tools_docker_proxy >nul 2>&1
docker rm tools_docker_proxy >nul 2>&1
docker rmi nofx-proxy:latest >nul 2>&1
docker network rm nofx-proxy-network >nul 2>&1
echo ✅ 清理完成
pause
goto menu

:health_check
echo.
echo 🚀 健康检查...
echo 测试端口连接...
netstat -an | findstr :8081
echo.
echo 测试HTTP响应...
curl -s http://localhost:8081/health
if %errorlevel% equ 0 (
    echo ✅ 服务响应正常
) else (
    echo ⚠️  服务可能未运行或端口被占用
)
pause
goto menu