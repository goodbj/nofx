@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 系统 - 完整部署脚本
echo ===============================================
echo.
echo 选择要部署的服务:
echo.
echo 1. 开发版 (Frontend: 3300, Backend: 8888, Chrome无头模式)
echo 2. 透明代理服务 (Port: 8081)
echo 3. 同时部署开发版和透明代理
echo 4. 仅部署开发版
echo 5. 显示分组服务 (nofx_dev_display - Chrome显示模式)
echo 6. 退出
echo.

set /p choice="请输入选项 (1-6): "

if "%choice%"=="1" (
    goto deploy_dev
) else if "%choice%"=="2" (
    goto deploy_proxy
) else if "%choice%"=="3" (
    goto deploy_both
) else if "%choice%"=="4" (
    goto deploy_dev_only
) else if "%choice%"=="5" (
    goto deploy_display
) else if "%choice%"=="6" (
    exit /b
) else (
    echo 无效选项，请重试
    pause
    goto start
)

:deploy_dev
echo.
echo 启动开发版服务...
call start_nofx_dev_docker.bat
goto end

:deploy_proxy
echo.
echo 启动透明代理服务...
call deploy_proxy_service.bat
goto end

:deploy_both
echo.
echo 启动开发版和透明代理服务...
echo 启动开发版服务...
call start_nofx_dev_docker.bat
echo.
echo 启动透明代理服务...
call deploy_proxy_service.bat
goto end

:deploy_dev_only
echo.
echo 启动开发版服务...
call start_nofx_dev_docker.bat
goto end

:deploy_display
echo.
echo 启动显示分组服务...
call start_nofx_dev_display.bat
goto end

:end
echo.
echo 部署完成！
pause