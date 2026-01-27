@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo Guardian Docker 一键启动脚本
echo ========================================
echo.
echo 重要提醒：请先确保Docker Desktop已启动并配置了镜像加速器
echo 参考: setup\Docker镜像加速配置指南.md
echo.

:menu
echo.
echo 请选择启动模式：
echo 1. 开发环境（含Guardian）- 推荐用于测试
echo 2. 生产环境（含Guardian）- 用于正式运行
echo 3. 检查Docker状态
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
echo 正在启动开发环境（含Guardian）...
echo.
echo 注意：首次启动可能需要较长时间下载镜像
echo 如果遇到网络问题，请参考快速启动指南
echo.
docker-compose -f docker-compose.dev.with-guardian.yml up -d --build
if %errorlevel% equ 0 (
    echo.
    echo ✅ 开发环境启动成功！
    echo.
    echo 服务状态：
    docker-compose -f docker-compose.dev.with-guardian.yml ps
    echo.
    echo Guardian日志（按Ctrl+C退出）：
    echo.
    docker logs -f nofx-guardian-dev
) else (
    echo.
    echo ❌ 开发环境启动失败，请检查错误信息
    echo.
    echo 可能的原因：
    echo - Docker未启动
    echo - 网络连接问题
    echo - 端口已被占用
    echo - 缺少必要的权限
    echo.
    echo 请参考: guardian\快速启动指南.md
)
goto pause_exit

:prod
echo.
echo 正在启动生产环境（含Guardian）...
echo.
echo 注意：首次启动可能需要较长时间下载镜像
echo 如果遇到网络问题，请参考快速启动指南
echo.
docker-compose -f docker-compose.with-guardian.yml up -d --build
if %errorlevel% equ 0 (
    echo.
    echo ✅ 生产环境启动成功！
    echo.
    echo 服务状态：
    docker-compose -f docker-compose.with-guardian.yml ps
    echo.
    echo Guardian日志（按Ctrl+C退出）：
    echo.
    docker logs -f nofx-guardian
) else (
    echo.
    echo ❌ 生产环境启动失败，请检查错误信息
    echo.
    echo 可能的原因：
    echo - Docker未启动
    echo - 网络连接问题
    echo - 端口已被占用
    echo - 缺少必要的权限
    echo.
    echo 请参考: guardian\快速启动指南.md
)
goto pause_exit

:check
echo.
echo 检查Docker状态...
docker --version
echo.
docker-compose --version
echo.
echo 当前运行的容器：
docker ps -a
goto pause_exit

:exit
echo.
echo 退出脚本...
goto pause_exit

:pause_exit
echo.
pause