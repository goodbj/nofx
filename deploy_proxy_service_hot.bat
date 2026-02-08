@echo off
REM 交易所代理服务热更新部署脚本
REM 用于部署支持热更新的透明代理服务到Docker容器中
REM 容器名称: tools_docker_proxy
REM 端口: 8081 (避免使用8080端口)

echo 正在部署支持热更新的交易所代理服务到Docker容器...

REM 切换到代理服务目录
cd /d "%~dp0\tools_docker_proxy"

echo 构建并启动支持热更新的代理服务容器...
docker compose -f docker-compose.hot.yml up -d --build

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ==================================
    echo 交易所代理服务热更新部署成功！
    echo.
    echo 容器名称: tools_docker_proxy
    echo 端口: 8081
    echo 功能: 处理受限地区交易所API请求，绕过地区限制
    echo 支持热更新
    echo ==================================
    echo.
    echo 检查服务状态...
    docker ps --filter "name=tools_docker_proxy"
    echo.
    echo 测试健康检查端点...
    curl http://localhost:8081/health
) else (
    echo.
    echo 错误: 部署失败，请检查错误信息
    pause
    exit /b 1
)

pause