@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🔄 热更新交易所代理服务
echo ========================================
echo.

echo 正在重新构建和部署服务...
echo.

REM 停止当前容器
docker-compose -f docker-compose.basic.yml down

REM 重新构建和启动
docker-compose -f docker-compose.basic.yml up -d --build

if %errorlevel% equ 0 (
    echo.
    echo ✅ 服务已重新部署
    echo 🌐 访问地址: http://localhost:8081
    echo 🏷️  端口: 8081
    echo 🔄 健康检查: http://localhost:8081/health
    echo.
    echo 容器状态:
    docker ps --filter "name=tools_docker_proxy"
) else (
    echo ❌ 部署失败
)

echo.
echo ========================================
echo 热更新完成
echo ========================================
echo.
pause