@echo off
chcp 65001 >nul
setlocal

echo 启动 NOFX 前端服务器...
echo ==============================

REM 检查是否已设置环境变量，如果没有则使用默认值
if "%FRONTEND_PORT%"=="" set FRONTEND_PORT=3300
if "%SERVER_PORT%"=="" set SERVER_PORT=8888

echo 正在启动前端服务器，端口: %FRONTEND_PORT%
echo 后端API地址将设置为: http://localhost:%SERVER_PORT%
echo.

REM 导航到前端目录
cd /d "e:\AI\nofx\web"

REM 检查是否已安装依赖
if not exist "node_modules" (
    echo 检测到首次运行，正在安装依赖...
    npm install
    if errorlevel 1 (
        echo.
        echo 错误: 依赖安装失败
        pause
        exit /b 1
    )
    echo 依赖安装完成
    echo.
)

REM 设置环境变量并启动前端
echo 启动命令: npm run dev -- --port %FRONTEND_PORT%
set PORT=%FRONTEND_PORT% && set VITE_API_BASE_URL=http://localhost:%SERVER_PORT% && npm run dev

if errorlevel 1 (
    echo.
    echo 错误: 前端服务器启动失败
    pause
    exit /b 1
)

pause