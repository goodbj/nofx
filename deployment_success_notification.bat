@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版-测试网页弹出功能版 部署成功！
echo ===============================================
echo.
echo 🎉 部署已完成，服务正在运行！
echo.
echo 服务状态：
echo   ✅ 后端容器: nofx-dev-backend-display
echo   ✅ 状态: 正在运行
echo   ✅ 端口: 8888
echo.
echo 功能特性：
echo   ✅ Chrome非无头模式（会弹窗）- 已启用
echo   ✅ 代码热更新 - 已启用
echo   ✅ Cookie持久化 - 已启用（./guardian/chrome_profile/）
echo   ✅ 数据库持久化 - 已启用（./data/data.db）
echo.
echo 访问地址：
echo   🔗 后端API: http://localhost:8888
echo.
echo 目录结构：
echo   📁 项目根目录: E:\AI\nofx_dev
echo   📁 数据库文件: E:\AI\nofx_dev\data\data.db
echo   📁 Chrome配置: E:\AI\nofx_dev\guardian\chrome_profile\
echo.
echo 控制命令：
echo   🛑 停止服务: docker compose -f docker-compose.dev.backend.only.display.yml down
echo   🔄 查看日志: docker compose -f docker-compose.dev.backend.only.display.yml logs -f
echo   🔍 检查状态: docker ps
echo.
echo 提示：
echo   • 请确保VcXsrv正在运行（DISPLAY环境变量已配置）
echo   • 当您在应用中触发浏览器自动化时，Chrome窗口将会弹出
echo   • 修改本地代码会自动更新到容器中（热更新）
echo   • 浏览器Cookie和数据已持久化保存
echo.
echo ===============================================
echo    部署成功！您可以开始使用了。
echo ===============================================
pause