@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版-测试网页弹出功能版 部署完成通知
echo ===============================================
echo.
echo 构建仍在进行中，这可能需要几分钟时间...
echo.
echo 当前状态：
echo ✅ Docker Compose配置文件已创建
echo ✅ Dockerfile已配置Chrome显示支持
echo ✅ 目录挂载已设置（支持热更新和Cookie持久化）
echo.
echo 服务详情：
echo - 后端容器名: nofx-dev-backend-display
echo - 后端端口: 8888
echo - Chrome模式: 非无头模式（会弹窗）
echo - 热更新: 已启用
echo - Cookie持久化: 已启用（./guardian/chrome_profile/）
echo - 数据库持久化: 已启用（./data/data.db）
echo.
echo 请继续等待构建完成...
echo.
echo 您可以打开另一个终端窗口检查构建进度：
echo   docker ps
echo 或查看日志：
echo   docker compose -f docker-compose.dev.backend.only.display.yml logs -f
echo.
echo ===============================================
echo    部署正在进行中...
echo ===============================================
pause