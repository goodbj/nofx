@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 透明代理终极修复脚本
echo ===============================================
echo.

echo 1. 停止所有相关服务...
docker compose -f docker-compose.dev.yml down 2>nul
docker compose -f tools_docker_proxy\docker-compose.unified.yml down 2>nul

echo.
echo 2. 停止现有的代理服务...
docker stop nofx-proxy 2>nul
docker rm nofx-proxy 2>nul

echo.
echo 3. 清理网络...
docker network rm nofx-unified-network nofx-docker-proxy_nofx-proxy-network 2>nul

echo.
echo 4. 创建新的统一网络...
docker network create nofx-unified-network

echo.
echo 5. 启动我们的代理服务（tools_docker_proxy）...
cd tools_docker_proxy
docker compose -f docker-compose.unified.yml up -d --build
cd ..

echo.
echo 6. 启动主服务...
docker compose -f docker-compose.dev.yml up -d --build

echo.
echo 7. 等待服务启动...
timeout /t 15 /nobreak >nul

echo.
echo 8. 验证网络配置...
echo.
echo --- 网络连接检查 ---
docker exec nofx-dev-backend ping -c 3 tools_docker_proxy
echo.
echo --- 代理服务检查 ---
curl -s http://localhost:8081/health
echo.
echo --- 后端服务检查 ---
curl -s http://localhost:8888/api/health

echo.
echo 9. 检查容器网络归属...
echo.
echo --- 后端容器网络 ---
docker inspect nofx-dev-backend | Select-String -Pattern "Networks" -Context 3
echo.
echo --- 代理容器网络 ---
docker inspect tools_docker_proxy | Select-String -Pattern "Networks" -Context 3

echo.
echo ===============================================
echo    修复完成！
echo ===============================================
echo.
echo 现在应该可以正常访问透明代理了
echo 请测试前端功能验证是否正常
echo.
pause