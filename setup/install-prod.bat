@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 稳定版 Docker 部署脚本
echo ========================================
echo.

REM 检查Docker是否安装
echo [1/4] 检查Docker安装状态...
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
echo [2/4] 检查项目目录...
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

REM 检查是否有对应的docker-compose文件
if not exist "..\docker-compose.stable.yml" (
    echo 错误: docker-compose.stable.yml 文件不存在
    echo 请确认您有稳定的生产配置文件
    pause
    exit /b 1
)

REM 构建稳定版服务
echo [3/4] 构建稳定版服务...
docker-compose -f ../docker-compose.stable.yml build
if %errorlevel% neq 0 (
    echo 错误: 生产服务构建失败
    pause
    exit /b 1
)
echo 生产服务构建完成
echo.

REM 启动稳定版服务
echo [4/4] 启动稳定版服务...
docker-compose -f ../docker-compose.stable.yml up -d
if %errorlevel% neq 0 (
    echo 错误: 稳定版服务启动失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo    稳定版构建和部署完成！
echo ========================================
echo.
echo 服务信息:
echo   - 请参考您的 docker-compose.stable.yml 配置文件
echo   - 数据库文件: E:\AI\nofx\data\data.db
echo.
echo 检查服务状态:
echo   docker-compose -f ../docker-compose.stable.yml ps
echo.
echo 查看日志:
echo   docker-compose -f ../docker-compose.stable.yml logs
echo ========================================
echo.

REM 清理构建缓存
echo [清理] 正在清理构建缓存...
go clean -cache -modcache 2>nul
echo 构建缓存清理完成
echo.

pause