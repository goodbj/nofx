@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX Dev Display 分组服务停止脚本
echo ===============================================
echo.

echo 停止 nofx_dev_display 分组服务...
docker stop nofx-dev-display-backend nofx-dev-display-frontend
docker rm nofx-dev-display-backend nofx-dev-display-frontend

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    NOFX Dev Display 分组服务已成功停止
    echo ===============================================
    echo.
) else (
    echo.
    echo 警告: 停止服务时出现错误
    echo.
)

echo 检查剩余服务:
docker ps | findstr nofx-dev-display
echo.
echo 数据目录状态:
echo - Chrome数据: E:\AI\nofx_Dev\data\browser_data
echo - 数据库文件: E:\AI\nofx_Dev\data\data.db
echo.
pause