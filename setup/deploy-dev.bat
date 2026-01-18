@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 开发版部署脚本
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
echo 正在启动开发版服务...
docker-compose -f docker-compose.dev.watch.yml up -d

echo.
echo 等待服务启动...
timeout /t 5 /nobreak >nul

echo.
echo 检查容器状态...
docker-compose -f docker-compose.dev.watch.yml ps

echo.
echo ========================================
echo    开发版部署完成
echo    前端访问地址: http://localhost:3300
echo    后端访问地址: http://localhost:8888
echo    支持代码热更新
echo ========================================
echo.

pause