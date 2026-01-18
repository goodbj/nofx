@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 双系统状态检查脚本
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

echo ========================================
echo    稳定版系统状态 (E:\AI\nofx)
echo ========================================
if exist "E:\AI\nofx" (
    cd /d "E:\AI\nofx"
    echo 当前目录: !CD!
    docker-compose -f docker-compose.stable.yml ps
    echo.
) else (
    echo 警告: E:\AI\nofx 目录不存在
    echo.
)

echo ========================================
echo    开发版系统状态 (E:\AI\nofx_Dev)
echo ========================================
if exist "E:\AI\nofx_Dev" (
    cd /d "E:\AI\nofx_Dev"
    echo 当前目录: !CD!
    docker-compose -f docker-compose.dev.watch.yml ps
    echo.
) else (
    echo 警告: E:\AI\nofx_Dev 目录不存在
    echo.
)

echo ========================================
echo    所有NOFX相关容器
echo ========================================
docker ps --filter "name=nofx"

echo.
echo ========================================
echo    服务访问地址:
echo    稳定版前端: http://localhost:3000
echo    稳定版后端: http://localhost:8080
echo    开发版前端: http://localhost:3300
echo    开发版后端: http://localhost:8888
echo ========================================
echo.

pause