@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🚀 启动交易所代理服务 (Docker + 热更新)
echo ========================================
echo.

echo 服务信息:
echo   - 容器名称: tools_docker_proxy
echo   - 端口: 8081 (避免使用8080端口)
echo   - 功能: 处理受限地区交易所API请求，绕过地区限制
echo   - 状态: 运行中
echo.

echo 正在启动Docker容器...
docker-compose -f docker-compose.basic.yml up -d --build

if %errorlevel% equ 0 (
    echo.
    echo ✅ 交易所代理服务已启动
    echo 🌐 访问地址: http://localhost:8081
    echo 🏷️  端口: 8081
    echo 🔄 健康检查: http://localhost:8081/health
    echo.
    echo 容器状态:
    docker ps --filter "name=tools_docker_proxy"
) else (
    echo ❌ 启动失败
    exit /b 1
)

echo.
echo ========================================
echo 📋 部署完成摘要:
echo   • 服务已部署到Docker容器
echo   • 使用端口8081 (非8080端口)
echo   • 容器名称: tools_docker_proxy
echo   • 支持热更新 (通过重新构建镜像实现)
echo   • 功能: 处理受限地区交易所API请求
echo ========================================
echo.
pause