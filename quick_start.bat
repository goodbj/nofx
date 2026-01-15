@echo off
chcp 65001 >nul
setlocal

echo NOFX 开发版快速启动
echo.

REM 检查 docker 是否可用
docker --version >nul 2>&1
if errorlevel 1 (
    echo 错误：未找到 Docker，请先安装 Docker Desktop
    pause
    exit /b 1
)

echo 请选择启动模式：
echo 1. 启动后端服务 (端口 8888)
echo 2. 启动前端服务 (端口 3300)  
echo 3. 启动全部服务
echo 4. 启动图形界面启动器
echo 5. 停止所有服务
echo.

set /p choice="请输入选择 (1-5): "

if "%choice%"=="1" (
    echo 启动后端服务...
    docker compose -f docker-compose.dev.watch.yml up nofx-dev-watch
) else if "%choice%"=="2" (
    echo 启动前端服务...
    docker compose -f docker-compose.dev.watch.yml up nofx-frontend-dev-watch
) else if "%choice%"=="3" (
    echo 启动全部服务...
    docker compose -f docker-compose.dev.watch.yml up
) else if "%choice%"=="4" (
    echo 启动图形界面启动器...
    if exist "launcher\\launcher_complete.py" (
        python launcher\\launcher_complete.py
    ) else (
        echo 未找到启动器脚本
    )
) else if "%choice%"=="5" (
    echo 停止所有服务...
    docker compose -f docker-compose.dev.watch.yml down
) else (
    echo 无效选择
)

pause