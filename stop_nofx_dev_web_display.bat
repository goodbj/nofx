@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    停止 NOFX 开发版-测试网页弹出功能版
echo ===============================================
echo.

echo 正在停止服务...
docker compose -f docker-compose.dev.web.display.yml down

echo.
echo 清理构建缓存...
docker builder prune -f

echo.
echo 服务已停止
echo.
pause