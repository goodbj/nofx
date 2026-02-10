@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX Dev Display 分组服务启动脚本
echo ===============================================
echo.

echo 检查并停止现有服务...
docker stop nofx-dev-display-backend nofx-dev-display-frontend 2>nul
docker rm nofx-dev-display-backend nofx-dev-display-frontend 2>nul

echo.
echo 启动后端服务 (nofx_dev_display 分组)...
echo - 端口: 8888
echo - 浏览器显示: 启用 (CHROME_HEADLESS=false)
echo - 热更新: 启用
echo - 数据持久化: E:\AI\nofx_Dev\storage 和 E:\AI\nofx_Dev\data
docker run -d --name nofx-dev-display-backend -p 8888:8888 -e CHROME_HEADLESS=false -e CHROME_BIN=/usr/bin/chromium-browser -e PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true -v "E:/AI/nofx_Dev/storage:/app/storage" -v "E:/AI/nofx_Dev/data:/app/data" nofx_dev-nofx-dev-backend-display:latest

if %errorlevel% neq 0 (
  echo 错误: 后端服务启动失败
  pause
  exit /b 1
)

echo.
echo 等待后端服务启动...
timeout /t 15 >nul

echo.
echo 启动前端服务 (nofx_dev_display 分组)...
echo - 端口: 3300
echo - 热更新: 启用
echo - API连接: http://host.docker.internal:8888
docker run -d --name nofx-dev-display-frontend -p 3300:3300 --link nofx-dev-display-backend -e NODE_ENV=development -e VITE_API_TARGET=http://host.docker.internal:8888 nofx_dev-nofx-dev-frontend-display:latest

if %errorlevel% neq 0 (
  echo 错误: 前端服务启动失败
  pause
  exit /b 1
)

echo.
echo ===============================================
echo    NOFX Dev Display 分组服务启动完成！
echo ===============================================
echo.
echo 服务状态:
docker ps | findstr nofx-dev-display
echo.
echo 访问地址:
echo   前端界面: http://localhost:3300
echo   后端API:  http://localhost:8888
echo.
echo 功能特性:
echo   - Chrome浏览器显示模式 (会弹窗)
echo   - 前端热更新 (Vite开发服务器)
echo   - 后端热更新 (Air工具)
echo   - 数据持久化 (Cookie、数据库)
echo   - 容器间通信正常
echo.
echo 数据目录:
echo   - Chrome数据: E:\AI\nofx_Dev\storage
echo   - 数据库文件: E:\AI\nofx_Dev\data\data.db
echo.
pause