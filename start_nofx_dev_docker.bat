@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版 Docker 部署启动脚本
echo ===============================================
echo.
echo 服务信息：
echo   前端: nofx-dev-frontend (端口 3300)
echo   后端: nofx-dev-backend (端口 8888)
echo   功能: Chrome无头模式（不弹窗）
echo   Chrome用户数据储存目录：E:\AI\nofx_Dev\data\browser_data
echo   数据库目录：E:\AI\nofx_Dev\data\data.db
echo.
echo 正在启动服务...
echo.

:: 检查Docker是否已安装并运行
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误：Docker未安装或未运行，请先启动Docker Desktop
    pause
    exit /b 1
)

:: 检查Docker Compose是否可用
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误：Docker Compose不可用
    pause
    exit /b 1
)

:: 启动服务
echo 启动nofx-dev开发版服务...
docker compose -f docker-compose.nofx-dev.yml up -d

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    NOFX 开发版服务已成功启动！
    echo ===============================================
    echo.
    echo 服务状态：
    docker compose -f docker-compose.nofx-dev.yml ps
    echo.
    echo 访问地址：
    echo   前端界面: http://localhost:3300
    echo   后端API: http://localhost:8888
    echo.
    echo Chrome浏览器将在后台以无头模式运行
    echo Chrome用户数据存储在: E:\AI\nofx_Dev\data\browser_data
    echo 数据库文件位置: E:\AI\nofx_Dev\data\data.db
    echo.
) else (
    echo.
    echo 错误：启动服务时出现问题
    echo 请检查错误信息并重试
)

pause