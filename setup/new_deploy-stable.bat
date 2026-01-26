@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 稳定版部署脚本
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

echo 切换到setup目录...
cd /d "%~dp0"

echo 当前目录: !CD!

echo.
echo 正在启动稳定版服务...
docker-compose -f docker-compose.stable.yml up -d

echo.
echo 等待服务启动...
timeout /t 5 /nobreak >nul

echo.
echo 检查容器状态...
docker-compose -f docker-compose.stable.yml ps

echo.
echo ========================================
echo    稳定版部署完成
echo    前端访问地址: http://localhost:3000
echo    后端访问地址: http://localhost:8080
echo ========================================
echo.

REM 清理构建缓存
echo [清理] 正在清理构建缓存...
go clean -cache -modcache 2>nul
echo 构建缓存清理完成
echo.

pause