@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 快速部署向导
echo ========================================
echo.
echo 请选择要部署的版本:
echo.
echo   [1] 开发版 (支持热更新)
echo   [2] 稳定版 (生产环境)
echo   [3] 查看当前部署状态
echo   [4] 清理环境
echo.
echo   [0] 退出
echo.

set /p choice="请输入选项编号 (0-4): "

if "%choice%"=="1" (
    call install-dev.bat
) else if "%choice%"=="2" (
    call install-prod.bat
) else if "%choice%"=="3" (
    call verify-deployment.bat
) else if "%choice%"=="4" (
    call cleanup.bat
) else if "%choice%"=="0" (
    exit /b 0
) else (
    echo 无效选项，请重新选择
    timeout /t 2
    cls
    goto :start
)

pause