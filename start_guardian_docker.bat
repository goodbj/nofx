@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo 启动含Guardian的NOFX Docker环境
echo ========================================

echo.
echo 请选择启动模式：
echo 1. 生产环境 (with-guardian)
echo 2. 开发环境 (dev.with-guardian)
echo 3. 退出
echo.

set /p choice="请输入选择 (1-3): "

if "%choice%"=="1" goto production
if "%choice%"=="2" goto development
if "%choice%"=="3" goto exit

echo 无效选择，退出...
goto exit

:production
echo.
echo 正在启动生产环境（含Guardian）...
docker-compose -f docker-compose.with-guardian.yml up -d
if %errorlevel% equ 0 (
    echo ✅ 生产环境启动成功
) else (
    echo ❌ 生产环境启动失败
)
goto status_prod

:development
echo.
echo 正在启动开发环境（含Guardian）...
docker-compose -f docker-compose.dev.with-guardian.yml up -d
if %errorlevel% equ 0 (
    echo ✅ 开发环境启动成功
) else (
    echo ❌ 开发环境启动失败
)
goto status_dev

:status_prod
echo.
echo 检查服务状态...
timeout /t 5 /nobreak >nul
docker-compose -f docker-compose.with-guardian.yml ps
goto pause_exit

:status_dev
echo.
echo 检查服务状态...
timeout /t 5 /nobreak >nul
docker-compose -f docker-compose.dev.with-guardian.yml ps
goto pause_exit

:exit
echo.
echo 退出脚本...
goto pause_exit

:pause_exit
echo.
pause