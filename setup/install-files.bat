@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 部署文件安装脚本
echo ========================================
echo.

echo 此脚本将部署文件复制到稳定版和开发版目录
echo.

:: 检查稳定版目录是否存在
if exist "E:\AI\nofx" (
    echo 复制文件到稳定版目录 (E:\AI\nofx)...
    if not exist "E:\AI\nofx\setup" mkdir "E:\AI\nofx\setup"
    copy /Y "%~dp0docker-compose.stable.yml" "E:\AI\nofx\setup\"
    copy /Y "%~dp0deploy-stable.bat" "E:\AI\nofx\setup\"
    echo.
) else (
    echo 警告: E:\AI\nofx 目录不存在，跳过稳定版文件复制
    echo.
)

:: 检查开发版目录是否存在
if exist "E:\AI\nofx_Dev" (
    echo 复制文件到开发版目录 (E:\AI\nofx_Dev)...
    if not exist "E:\AI\nofx_Dev\setup" mkdir "E:\AI\nofx_Dev\setup"
    copy /Y "%~dp0docker-compose.dev.watch.yml" "E:\AI\nofx_Dev\setup\"
    copy /Y "%~dp0deploy-dev.bat" "E:\AI\nofx_Dev\setup\"
    echo.
) else (
    echo 警告: E:\AI\nofx_Dev 目录不存在，跳过开发版文件复制
    echo.
)

echo ========================================
echo    文件安装完成
echo ========================================
echo.

pause