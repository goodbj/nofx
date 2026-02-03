@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ================================
echo 币安代理服务安装脚本
echo ================================

:: 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ 错误: 需要管理员权限运行此脚本
    echo 请右键点击此批处理文件，选择"以管理员身份运行"
    pause
    exit /b 1
)

:: 设置工作目录
cd /d "%~dp0"

:: 检查Docker是否安装
echo 检查Docker环境...
docker --version >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ 错误: 未检测到Docker环境
    echo 请先安装Docker Desktop: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo ✅ Docker环境正常

:: 检查端口占用
echo 检查端口8081占用情况...
netstat -an | findstr :8081 >nul
if %errorLevel% equ 0 (
    echo ⚠️  端口8081已被占用，正在检查占用进程...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081 ^| findstr LISTENING') do (
        set pid=%%a
        for /f "tokens=1" %%b in ('tasklist /fi "pid eq !pid!" /fo csv /nh 2^>nul') do (
            echo   占用进程: %%~b (PID: !pid!)
        )
    )
    echo.
    echo 请选择操作:
    echo 1. 停止占用进程并继续安装
    echo 2. 退出安装
    choice /c 12 /m "请输入选择(1/2)"
    if errorlevel 2 (
        echo 已取消安装
        exit /b 0
    )
    if errorlevel 1 (
        echo 正在停止占用进程...
        for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081 ^| findstr LISTENING') do (
            taskkill /f /pid %%a >nul 2>&1
        )
        timeout /t 2 /nobreak >nul
    )
)

:: 构建Docker镜像
echo.
echo ================================
echo 开始构建币安代理服务镜像...
echo ================================

cd api\binance_proxy
if exist Dockerfile.simple (
    echo 使用简化版Dockerfile构建...
    docker build -t binance-proxy-service -f Dockerfile.simple ../..
) else (
    echo 使用默认Dockerfile构建...
    docker build -t binance-proxy-service .
)

if %errorLevel% neq 0 (
    echo ❌ 镜像构建失败
    cd ..\..
    pause
    exit /b 1
)

echo ✅ 镜像构建成功

:: 返回项目根目录
cd ..\..

:: 清理旧容器
echo.
echo 清理旧的代理服务容器...
docker stop binance-proxy-container >nul 2>&1
docker rm binance-proxy-container >nul 2>&1

:: 启动容器
echo.
echo 启动币安代理服务容器...
docker run -d -p 8081:8081 --name binance-proxy-container binance-proxy-service

if %errorLevel% neq 0 (
    echo ❌ 容器启动失败
    pause
    exit /b 1
)

echo ✅ 容器启动成功

:: 等待服务启动
echo.
echo 等待服务启动...
timeout /t 5 /nobreak >nul

:: 测试服务
echo.
echo 测试代理服务连接...
curl -s http://localhost:8081/health >nul 2>&1
if %errorLevel% equ 0 (
    echo ✅ 代理服务连接成功
    echo 服务地址: http://localhost:8081
    echo 健康检查: http://localhost:8081/health
) else (
    echo ⚠️  服务连接测试失败，请稍后手动检查
)

:: 显示容器状态
echo.
echo ================================
echo 容器运行状态:
echo ================================
docker ps | findstr binance-proxy

echo.
echo ================================
echo 安装完成!
echo ================================
echo 币安代理服务已成功部署
echo 镜像名称: binance-proxy-service
echo 容器名称: binance-proxy-container
echo 监听端口: 8081
echo.
echo 常用管理命令:
echo   查看日志: docker logs binance-proxy-container
echo   停止服务: docker stop binance-proxy-container
echo   启动服务: docker start binance-proxy-container
echo   重启服务: docker restart binance-proxy-container
echo   删除容器: docker rm binance-proxy-container
echo.
pause