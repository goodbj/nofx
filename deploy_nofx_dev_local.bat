@echo off
echo ==================================================
echo NOFX_DEV 本地开发环境部署脚本
echo ==================================================
echo 开始时间: %date% %time%
echo.

REM 检查必要目录
echo [1/6] 检查和创建必要目录...
if not exist "E:\AI\nofx_dev" (
    echo 创建主目录: E:\AI\nofx_dev
    mkdir "E:\AI\nofx_dev"
)

if not exist "E:\AI\nofx_dev\data" (
    echo 创建数据目录: E:\AI\nofx_dev\data
    mkdir "E:\AI\nofx_dev\data"
)

if not exist "E:\AI\nofx_dev\storage" (
    echo 创建存储目录: E:\AI\nofx_dev\storage
    mkdir "E:\AI\nofx_dev\storage"
)

echo 目录检查完成
echo.

REM 停止可能正在运行的服务
echo [2/6] 停止现有服务...
taskkill /f /im go.exe 2>nul
taskkill /f /im node.exe 2>nul
echo 现有服务已停止
echo.

REM 设置环境变量
echo [3/6] 设置环境变量...
set NOFX_DEV_PATH=E:\AI\nofx_dev
set DATABASE_PATH=E:\AI\nofx_dev\data\data.db
set STORAGE_PATH=E:\AI\nofx_dev\storage
set CHROME_HEADLESS=false
echo 环境变量设置完成
echo.

REM 启动后端服务
echo [4/6] 启动后端服务 (端口 8888)...
cd /d "E:\AI\nofx_Dev"
start "NOFX Backend Service" cmd /c "go run main.go --port=8888 --display=true --hot-reload=true"
timeout /t 5 /nobreak >nul
echo 后端服务启动中...
echo.

REM 启动前端服务
echo [5/6] 启动前端服务 (端口 3300)...
cd /d "E:\AI\nofx_Dev\web"
if exist "node_modules" (
    echo Node modules 已存在
) else (
    echo 安装前端依赖...
    npm install
)
start "NOFX Frontend Service" cmd /c "npm run dev -- --host --port 3300"
timeout /t 5 /nobreak >nul
echo 前端服务启动中...
echo.

REM 验证服务
echo [6/6] 验证服务状态...
timeout /t 10 /nobreak >nul
echo.
echo 当前运行的进程:
tasklist | findstr "go.exe\|node.exe"
echo.
echo 测试端口连接:
netstat -an | findstr "8888\|3300"
echo.

echo ==================================================
echo 部署完成!
echo ==================================================
echo 服务信息:
echo - 后端服务: http://localhost:8888 (端口 8888)
echo - 前端服务: http://localhost:3300 (端口 3300)
echo - 浏览器显示: 已启用 (Chrome非无头模式)
echo - 热更新: 已启用
echo - 数据库路径: E:\AI\nofx_dev\data\data.db
echo - Cookie持久化: E:\AI\nofx_dev\storage
echo.
echo 管理命令:
echo - 查看后端日志: 查看 "NOFX Backend Service" 命令窗口
echo - 查看前端日志: 查看 "NOFX Frontend Service" 命令窗口
echo - 停止服务: 运行 stop_nofx_dev.bat
echo ==================================================
echo 完成时间: %date% %time%
pause