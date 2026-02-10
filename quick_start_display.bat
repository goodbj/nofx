@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 显示分组服务快速启动
echo    标签分组: nofx_dev_display
echo ===============================================
echo.
echo 一键启动开发版显示分组服务...
echo.

:: 直接调用显示分组启动脚本
call start_nofx_dev_display.bat

echo.
echo 显示分组服务启动完成！
echo 访问地址:
echo   前端: http://localhost:3300
echo   后端: http://localhost:8888
echo.