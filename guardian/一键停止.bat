@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo Guardian Docker 一键停止脚本
echo ========================================
echo.

:menu
echo.
echo 请选择停止模式：
echo 1. 停止开发环境（含Guardian）
echo 2. 停止生产环境（含Guardian）
echo 3. 检查运行状态
echo 4. 退出
echo.

set /p choice="请输入选择 (1-4): "

if "%choice%"=="1" goto dev
if "%choice%"=="2" goto prod
if "%choice%"=="3" goto check
if "%choice%"=="4" goto exit

echo 无效选择，请重新输入。
goto menu

:dev
echo.
echo 正在停止开发环境（含Guardian）...
docker-compose -f docker-compose.dev.with-guardian.yml down
if %errorlevel% equ 0 (
    echo ✅ 开发环境已停止
) else (
    echo ❌ 停止开发环境时出现问题
)
goto pause_exit

:prod
echo.
echo 正在停止生产环境（含Guardian）...
docker-compose -f docker-compose.with-guardian.yml down
if %errorlevel% equ 0 (
    echo ✅ 生产环境已停止
) else (
    echo ❌ 停止生产环境时出现问题
)
goto pause_exit

:check
echo.
echo 检查当前运行的服务...
echo.
echo 开发环境状态：
docker-compose -f docker-compose.dev.with-guardian.yml ps
echo.
echo 生产环境状态：
docker-compose -f docker-compose.with-guardian.yml ps
goto pause_exit

:exit
echo.
echo 退出脚本...
goto pause_exit

:pause_exit
echo.
pause