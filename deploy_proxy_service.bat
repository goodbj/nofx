@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 系统 - 透明代理服务 Docker 部署脚本
echo ===============================================
echo.
echo 正在启动透明代理服务...
echo.

:: 检查Docker是否已安装并运行
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误：Docker未安装或未运行，请先启动Docker Desktop
    pause
    exit /b 1
)

:: 检查Docker Compose是否可用
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误：Docker Compose不可用
    pause
    exit /b 1
)

:: 启动透明代理服务
echo 启动 NOFX 透明代理服务...
cd nofx-docker-proxy
docker compose up -d

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    NOFX 透明代理服务已成功启动！
    echo ===============================================
    echo.
    echo 服务状态：
    docker compose ps
    echo.
    echo 访问地址：
    echo   代理服务: http://localhost:8081
    echo   健康检查: http://localhost:8081/health
    echo.
    echo 提示：服务通过 X-Custom-API-URL 请求头接收目标API地址
    echo.
) else (
    echo.
    echo 错误：启动服务时出现问题
    echo 请检查错误信息并重试
)

pause