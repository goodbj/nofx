@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版-测试网页弹出功能版 部署脚本
echo ===============================================
echo.
echo 项目要求：
echo - 前端: nofx-dev-frontend-display (端口 3300)
echo - 后端: nofx-dev-backend-display (端口 8888) 
echo - 功能: Chrome非无头模式（会弹窗），用于测试浏览器自动化
echo - 支持热更新
echo - 浏览器Cookie持久化
echo - 数据库: E:\AI\nofx_dev\data\data.db
echo ===============================================
echo.

echo 1. 检查必要目录...
if not exist "data" (
    echo 创建 data 目录...
    mkdir data
)
if not exist "guardian" (
    echo 创建 guardian 目录...
    mkdir guardian
)
if not exist "guardian\chrome_profile" (
    echo 创建 chrome_profile 目录...
    mkdir guardian\chrome_profile
)
if not exist "data\data.db" (
    echo 初始化数据库文件...
    copy nul data\data.db >nul
)

echo.
echo 2. 停止现有服务...
docker compose -f docker-compose.dev.web.display.yml down 2>nul

echo.
echo 3. 构建并启动服务...
docker compose -f docker-compose.dev.web.display.yml up -d --build

echo.
echo 4. 等待服务启动...
timeout /t 30 /nobreak >nul

echo.
echo 5. 验证服务状态...
echo.
echo --- 后端服务状态 ---
docker ps --filter "name=nofx-dev-backend-display" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.
echo --- 前端服务状态 ---
docker ps --filter "name=nofx-dev-frontend-display" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo.
echo 6. 检查服务健康状态...
echo.
echo 测试后端API健康检查:
curl -s http://localhost:8888/api/health
echo.
echo 测试前端页面:
curl -s http://localhost:3300

echo.
echo ===============================================
echo    部署完成！
echo ===============================================
echo.
echo 访问地址：
echo   前端页面: http://localhost:3300
echo   后端API:  http://localhost:8888
echo   后端健康检查: http://localhost:8888/api/health
echo.
echo 功能特性：
echo ✅ Chrome非无头模式（会弹窗）- 用于测试浏览器自动化
echo ✅ 支持热更新 - 修改代码后自动重新加载
echo ✅ Cookie持久化 - 浏览器登录状态保存在 ./guardian/chrome_profile/
echo ✅ 数据库持久化 - 数据库文件位于 ./data/data.db
echo.
echo 使用说明：
echo - 启动后，当触发浏览器自动化功能时，Chrome窗口将弹出显示
echo - 修改后端代码后，Air工具会自动热更新
echo - 登录信息和浏览器配置将持久保存
echo - 数据库文件位于 E:\AI\nofx_dev\data\data.db
echo.
echo 如需停止服务，请运行：
echo   docker compose -f docker-compose.dev.web.display.yml down
echo.
pause