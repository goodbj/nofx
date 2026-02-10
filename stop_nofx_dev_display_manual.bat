@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX Dev Display 服务停止脚本
echo ===============================================
echo.

echo 停止 nofx_dev_display 分组服务...
docker stop nofx-dev-backend-display nofx-dev-frontend-display
docker rm nofx-dev-backend-display nofx-dev-frontend-display

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    服务已成功停止
    echo ===============================================
    echo.
) else (
    echo.
    echo 警告: 停止服务时出现错误
    echo.
)

echo 检查剩余服务:
docker ps | findstr nofx-dev
echo.
pause