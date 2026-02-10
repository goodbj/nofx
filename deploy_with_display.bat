@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 带浏览器显示的后端部署脚本
echo ===============================================
echo.

echo 1. 停止现有服务...
docker compose -f docker-compose.dev.yml down 2>nul
docker compose -f docker-compose.dev.display.grouped.yml down 2>nul

echo.
echo 2. 启动带显示功能的后端服务...
docker compose -f docker-compose.dev.display.grouped.yml up -d backend-display

echo.
echo 3. 等待后端服务启动...
timeout /t 15 /nobreak >nul

echo.
echo 4. 启动前端服务...
docker compose -f docker-compose.dev.display.grouped.yml up -d frontend-display

echo.
echo 5. 验证服务状态...
echo.
echo --- 后端服务状态 ---
docker ps --filter "name=nofx-dev-backend-display" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.
echo --- 前端服务状态 ---
docker ps --filter "name=nofx-dev-frontend-display" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo.
echo 6. 测试连接...
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
echo 说明：
echo - 后端现在支持浏览器显示功能
echo - 当触发浏览器自动化时，Chrome窗口将显示在宿主机上
echo - 前端通过 http://host.docker.internal:8890 连接到后端
echo.
echo 注意事项：
echo - 确保Windows已启用X11转发或使用本地显示
echo - 如果需要调试浏览器窗口，可以查看容器日志
echo.
pause