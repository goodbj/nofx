@echo off
echo ==================================================
echo NOFX_DEV 分组服务重新部署脚本
echo ==================================================
echo 开始时间: %date% %time%
echo.

REM 检查 Docker 环境
echo [1/8] 检查 Docker 环境...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker 未安装或未运行
    pause
    exit /b 1
)
echo Docker 环境正常
echo.

REM 停止现有服务
echo [2/8] 停止现有 nofx-dev 服务...
docker stop nofx-dev-display-backend nofx-dev-display-frontend 2>nul
docker rm nofx-dev-display-backend nofx-dev-display-frontend 2>nul
echo 现有服务已停止
echo.

REM 清理网络
echo [3/8] 清理网络配置...
docker network rm nofx-dev-display-network 2>nul
echo 网络清理完成
echo.

REM 检查并创建必要的目录
echo [4/8] 检查和创建必要目录...
if not exist "E:\AI\nofx_dev" (
    echo 创建主目录: E:\AI\nofx_dev
    mkdir "E:\AI\nofx_dev"
)

if not exist "E:\AI\nofx_dev\data" (
    echo 创建数据目录: E:\AI\nofx_dev\data
    mkdir "E:\AI\nofx_dev\data"
)

if not exist "E:\AI\nofx_dev\storage" (
    echo 创建存储目录: E:\AI\nofx_dev\storage
    mkdir "E:\AI\nofx_dev\storage"
)

echo 目录检查完成
echo.

REM 构建和启动服务
echo [5/8] 构建和启动服务...
echo 使用配置文件: docker-compose.nofx-dev-display-final.yml
docker-compose -f docker-compose.nofx-dev-display-final.yml up -d --build
if %errorlevel% neq 0 (
    echo 错误: 服务启动失败
    pause
    exit /b 1
)
echo 服务构建和启动完成
echo.

REM 等待服务启动
echo [6/8] 等待服务启动...
timeout /t 10 /nobreak >nul
echo.

REM 检查服务状态
echo [7/8] 检查服务状态...
echo.
echo 当前运行的容器:
docker ps --filter "name=nofx-dev-display"
echo.
echo 服务日志:
docker logs nofx-dev-display-backend --tail 20
echo.
docker logs nofx-dev-display-frontend --tail 10
echo.

REM 验证端口
echo [8/8] 验证端口访问...
echo.
echo 测试后端服务 (8888端口):
curl -s http://localhost:8888/health 2>nul && echo "✓ 后端服务正常" || echo "✗ 后端服务异常"
echo.
echo 测试前端服务 (3300端口):
curl -s http://localhost:3300 2>nul && echo "✓ 前端服务正常" || echo "✗ 前端服务异常"
echo.

echo ==================================================
echo 部署完成!
echo ==================================================
echo 服务信息:
echo - 后端服务: http://localhost:8888 (端口 8888)
echo - 前端服务: http://localhost:3300 (端口 3300)
echo - 浏览器显示: 已启用 (Chrome非无头模式)
echo - 热更新: 已启用
echo - 数据库路径: E:\AI\nofx_dev\data\data.db
echo - Cookie持久化: E:\AI\nofx_dev\storage
echo.
echo 服务管理命令:
echo - 查看日志: docker logs nofx-dev-display-backend
echo - 停止服务: docker-compose -f docker-compose.nofx-dev-display-final.yml down
echo - 重启服务: docker-compose -f docker-compose.nofx-dev-display-final.yml restart
echo ==================================================
echo 完成时间: %date% %time%
pause