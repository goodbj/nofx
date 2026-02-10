@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 带浏览器显示和热更新的后端部署脚本
echo ===============================================
echo.

echo 1. 停止现有服务...
docker compose -f docker-compose.dev.display.grouped.yml down 2>nul
docker compose -f docker-compose.dev.display.watch.yml down 2>nul

echo.
echo 2. 创建必要的目录...
if not exist "data" mkdir data
if not exist "guardian" mkdir guardian
if not exist "guardian\chrome_profile" mkdir guardian\chrome_profile

echo.
echo 3. 启动带显示和热更新功能的后端服务...
docker compose -f docker-compose.dev.display.watch.yml up -d --build backend-display-watch

echo.
echo 4. 等待后端服务启动...
timeout /t 20 /nobreak >nul

echo.
echo 5. 启动前端服务...
docker compose -f docker-compose.dev.display.watch.yml up -d frontend-display-watch

echo.
echo 6. 验证服务状态...
echo.
echo --- 后端服务状态 ---
docker ps --filter "name=nofx-dev-backend-display-watch" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.
echo --- 前端服务状态 ---
docker ps --filter "name=nofx-dev-frontend-display-watch" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo.
echo 7. 测试连接...
echo 测试后端API:
curl -s http://localhost:8890/api/health
echo.
echo 测试前端页面:
curl -s http://localhost:3301

echo.
echo ===============================================
echo    部署完成！
echo ===============================================
echo.
echo 访问地址：
echo   前端页面: http://localhost:3301
echo   后端API:  http://localhost:8890
echo   后端健康检查: http://localhost:8890/api/health
echo.
echo 功能特性：
echo ✅ 后端支持浏览器显示功能（Chrome窗口将在宿主机显示）
echo ✅ 支持代码热更新（修改Go代码后自动重新编译运行）
echo ✅ Cookie持久化（登录信息保存在 ./guardian/chrome_profile/）
echo ✅ 数据持久化（数据保存在 ./data/ 目录）
echo.
echo 使用说明：
echo - 当触发浏览器自动化时，Chrome窗口将显示在宿主机上
echo - 修改后端代码后，Air工具会自动重新编译和运行
echo - 登录信息和浏览器配置将持久保存
echo - 重要数据将保存在本地data目录中
echo.
pause