@echo off
chcp 65001 >nul
setlocal

echo 启动 NOFX 全栈应用程序...
echo ==============================

REM 设置默认端口
if "%SERVER_PORT%"=="" set SERVER_PORT=8888
if "%FRONTEND_PORT%"=="" set FRONTEND_PORT=3300

echo 后端端口: %SERVER_PORT%
echo 前端端口: %FRONTEND_PORT%
echo.

REM 启动后端服务器（在新窗口中）
echo 启动后端服务器...
start "NOFX 后端服务器" cmd /k "cd /d \"e:\AI\nofx\" && go run main.go"

timeout /t 5 /nobreak >nul

REM 启动前端服务器（在新窗口中）
echo 启动前端服务器...
start "NOFX 前端服务器" cmd /k "cd /d \"e:\AI\nofx\web\" && set PORT=%FRONTEND_PORT% && set VITE_API_BASE_URL=http://localhost:%SERVER_PORT% && npm run dev"

echo.
echo 应用程序已启动！
echo - 后端服务器将在新窗口中运行，地址: http://localhost:%SERVER_PORT%
echo - 前端服务器将在新窗口中运行，地址: http://localhost:%FRONTEND_PORT%
echo.
echo 注意: 请保持两个命令行窗口开启以维持服务运行
pause