@echo off
echo ==================================================
echo 交易所代理服务 Docker 部署脚本
echo ==================================================
echo 开始时间: %date% %time%
echo.

REM 检查 Docker 环境
echo [1/7] 检查 Docker 环境...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker 未安装或未运行
    pause
    exit /b 1
)
echo Docker 环境正常
echo.

REM 停止现有容器
echo [2/7] 停止现有代理服务容器...
docker stop tools_docker_proxy 2>nul
docker rm tools_docker_proxy 2>nul
echo 现有容器已停止
echo.

REM 清理网络
echo [3/7] 清理网络配置...
docker network rm nofx-proxy-network 2>nul
echo 网络清理完成
echo.

REM 设置环境变量
echo [4/7] 设置环境变量...
set PROXY_PORT=8081
echo 端口设置为: %PROXY_PORT%
echo.

REM 构建和启动服务
echo [5/7] 构建和启动代理服务...
cd /d "E:\AI\nofx_Dev\tools_docker_proxy"
echo 使用热更新配置文件: docker-compose.hot.yml
docker-compose -f docker-compose.hot.yml up -d --build
if %errorlevel% neq 0 (
    echo 错误: 服务启动失败
    pause
    exit /b 1
)
echo 服务构建和启动完成
echo.

REM 等待服务启动
echo [6/7] 等待服务启动...
timeout /t 10 /nobreak >nul
echo.

REM 检查服务状态
echo [7/7] 检查服务状态...
echo.
echo 当前运行的容器:
docker ps --filter "name=tools_docker_proxy"
echo.
echo 服务日志:
docker logs tools_docker_proxy --tail 20
echo.
echo 端口检查:
netstat -an | findstr :8081
echo.

echo ==================================================
echo 部署完成!
echo ==================================================
echo 服务信息:
echo - 容器名称: tools_docker_proxy
echo - 端口: 8081 (避免使用8080)
echo - 功能: 处理受限地区交易所API请求，绕过地区限制
echo - 热更新: 已启用
echo - 访问地址: http://localhost:8081
echo.
echo 管理命令:
echo - 查看日志: docker logs tools_docker_proxy
echo - 停止服务: docker stop tools_docker_proxy
echo - 重启服务: docker restart tools_docker_proxy
echo - 重新部署: 重新运行此脚本
echo ==================================================
echo 完成时间: %date% %time%
pause