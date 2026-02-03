@echo off
echo ========================================
echo    NOFX 开发版 Docker 部署脚本
echo ========================================
echo.
echo 正在启动 NOFX 开发版 (前端端口3300, 后端端口8888)
echo 功能: Chrome无头模式, 支持热更新
echo 数据库路径: E:\AI\nofx_dev\data\data.db
echo.
echo 按任意键开始部署...
pause >nul

docker-compose -f docker-compose.nofx-dev-chrome-hot.yml up --build -d

if %errorlevel% == 0 (
    echo.
    echo ========================================
    echo    NOFX 开发版部署成功!
    echo ========================================
    echo.
    echo 前端访问地址: http://localhost:3300
    echo 后端API地址: http://localhost:8888
    echo.
    echo Chrome无头模式已启用
    echo 热更新功能已激活
    echo 数据库文件位置: E:\AI\nofx_dev\data\data.db
    echo.
    echo 按 Ctrl+C 可停止服务
    docker-compose -f docker-compose.nofx-dev-chrome-hot.yml logs -f
) else (
    echo.
    echo ========================================
    echo    部署失败，请检查错误信息
    echo ========================================
    pause
)