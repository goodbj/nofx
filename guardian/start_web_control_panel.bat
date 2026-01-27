@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo 启动 Guardian Web 控制面板
echo ========================================

echo.
echo 正在启动 Guardian Web 控制面板...
echo 服务器将在 http://localhost:8085 运行
echo.

cd /d "e:\AI\nofx_Dev\guardian"

echo 正在编译并启动 Guardian 控制面板服务器...
go run web_controller.go

echo.
echo Guardian Web 控制面板已停止
pause