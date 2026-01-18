@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 部署状态验证脚本
echo ========================================
echo.

echo 检查Docker服务状态...
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker服务未运行
    echo 请启动Docker Desktop应用程序
    pause
    exit /b 1
)

echo Docker服务运行正常
echo.

echo 检查开发版服务状态...
echo.
docker-compose -f ../docker-compose.dev.watch.yml ps
echo.

echo 检查稳定版服务状态...
echo.
docker-compose -f ../docker-compose.stable.yml ps
echo.

echo 检查正在运行的容器...
echo.
docker ps --filter "name=nofx"
echo.

echo ========================================
echo    服务访问地址:
echo    前端界面: http://localhost:3300
echo    后端API: http://localhost:8888
echo ========================================
echo.

pause