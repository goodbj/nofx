@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 开发版 Docker 部署脚本
echo ========================================
echo.

REM 检查Docker是否安装
echo [1/5] 检查Docker安装状态...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker 未安装或未在PATH中
    echo 请先安装Docker Desktop并启动Docker服务
    pause
    exit /b 1
)

REM 检查Docker服务是否运行
docker ps >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker服务未运行
    echo 请启动Docker Desktop应用程序
    pause
    exit /b 1
)

echo Docker服务运行正常
echo.

REM 检查项目根目录
echo [2/5] 检查项目目录...
if not exist "..\.env" (
    echo 警告: .env文件不存在，将使用默认配置
    echo 请确保在项目根目录下配置好环境变量
)

REM 检查data目录
if not exist "..\data" (
    echo 创建数据目录: ..\data
    mkdir "..\data"
)

REM 检查data.db文件
if not exist "..\data\data.db" (
    echo 创建数据库文件: ..\data\data.db
    type nul > "..\data\data.db"
)

echo 项目目录检查完成
echo.

REM 构建后端服务
echo [3/5] 构建后端服务 (nofx-trading-dev-watch)...
docker-compose -f ../docker-compose.dev.watch.yml build nofx-trading-dev-watch
if %errorlevel% neq 0 (
    echo 错误: 后端服务构建失败
    pause
    exit /b 1
)
echo 后端服务构建完成
echo.

REM 构建前端服务
echo [4/5] 构建前端服务 (nofx-frontend-dev-watch)...
docker-compose -f ../docker-compose.dev.watch.yml build nofx-frontend-dev-watch
if %errorlevel% neq 0 (
    echo 错误: 前端服务构建失败
    pause
    exit /b 1
)
echo 前端服务构建完成
echo.

REM 启动服务
echo [5/5] 启动服务...
docker-compose -f ../docker-compose.dev.watch.yml up -d
if %errorlevel% neq 0 (
    echo 错误: 服务启动失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo    构建和部署完成！
echo ========================================
echo.
echo 服务信息:
echo   - 前端界面: http://localhost:3300
echo   - 后端API: http://localhost:8888
echo   - 数据库文件: E:\AI\nofx_Dev\data\data.db
echo.
echo 检查服务状态:
echo   docker-compose -f ../docker-compose.dev.watch.yml ps
echo.
echo 查看日志:
echo   docker logs nofx-trading-dev-watch
echo   docker logs nofx-frontend-dev-watch
echo ========================================
echo.

pause