@echo off
echo ==================================================
echo 停止 NOFX_DEV 服务
echo ==================================================
echo.

echo 停止后端服务...
taskkill /f /im go.exe 2>nul
echo 后端服务已停止

echo 停止前端服务...
taskkill /f /im node.exe 2>nul
echo 前端服务已停止

echo.
echo 所有服务已停止
echo ==================================================
pause