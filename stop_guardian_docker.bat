@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo 停止含Guardian的NOFX Docker环境
echo ========================================

echo.
echo 请选择停止模式：
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
echo 正在停止生产环境（含Guardian）...
docker-compose -f docker-compose.with-guardian.yml down
if %errorlevel% equ 0 (
    echo ✅ 生产环境停止成功
) else (
    echo ❌ 生产环境停止失败
)
goto pause_exit

:development
echo.
echo 正在停止开发环境（含Guardian）...
docker-compose -f docker-compose.dev.with-guardian.yml down
if %errorlevel% equ 0 (
    echo ✅ 开发环境停止成功
) else (
    echo ❌ 开发环境停止失败
)
goto pause_exit

:exit
echo.
echo 退出脚本...
goto pause_exit

:pause_exit
echo.
pause