@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX Dev 分组服务启动脚本
echo ===============================================
echo.

echo 检查并停止现有服务...
docker stop nofx-dev-backend nofx-dev-frontend 2>nul

echo.
echo 启动后端服务 (nofx_dev 分组)...
docker run -d --name nofx-dev-backend -p 8888:8888 -e CHROME_HEADLESS=false -e CHROME_BIN=/usr/bin/chromium-browser -e PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true -v "%CD%/storage:/app/storage" -v "%CD%/data:/app/data" nofx_dev-nofx-dev-backend-display:latest

if %errorlevel% neq 0 (
  echo 错误: 后端服务启动失败
  pause
  exit /b 1
)

echo.
echo 等待后端服务启动...
timeout /t 10 >nul

echo.
echo 启动前端服务 (nofx_dev 分组)...
docker run -d --name nofx-dev-frontend -p 3300:3300 --link nofx-dev-backend -e NODE_ENV=development -e VITE_API_TARGET=http://host.docker.internal:8888 nofx_dev-nofx-dev-frontend-display:latest

if %errorlevel% neq 0 (
  echo 错误: 前端服务启动失败
  pause
  exit /b 1
)

echo.
echo ===============================================
echo    NOFX Dev 分组服务启动完成！
echo ===============================================
echo.
echo 服务状态:
docker ps | findstr nofx-dev
echo.
echo 访问地址:
echo   前端界面: http://localhost:3300
echo   后端API:  http://localhost:8888
echo.
echo 功能特性:
echo   - Chrome浏览器显示模式 (CHROME_HEADLESS=false)
echo   - 前端热更新 (Vite开发服务器)
echo   - 后端热更新 (Air工具)
echo   - 数据持久化 (data/browser_data/ 和 data/ 目录)
echo.
pause