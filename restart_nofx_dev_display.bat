@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版显示分组服务 - 重启脚本
echo    标签分组: nofx_dev_display
echo ===============================================
echo.

:: 检查Docker是否运行
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker未运行，请先启动Docker Desktop
    pause
    exit /b 1
)

echo 正在重启 nofx_dev_display 分组服务...
echo.

:: 停止服务
echo 停止现有服务...
docker compose -f docker-compose.nofx-dev-display.yml down

:: 启动服务
echo 启动服务...
docker compose -f docker-compose.nofx-dev-display.yml up -d

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    nofx_dev_display 分组服务重启成功!
    echo ===============================================
    echo.
    echo 服务状态:
    docker compose -f docker-compose.nofx-dev-display.yml ps
    echo.
    echo 访问地址:
    echo   前端界面: http://localhost:3300
    echo   后端API:  http://localhost:8888
    echo.
) else (
    echo.
    echo 错误: 服务重启失败
    echo 请检查错误信息
)

pause