@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 透明代理网络修复脚本
echo ===============================================
echo.

echo 停止当前服务...
docker compose -f docker-compose.dev.yml down
docker compose -f tools_docker_proxy\docker-compose.yml down

echo.
echo 创建统一网络...
docker network create nofx-unified-network 2>nul

echo.
echo 重新启动服务到统一网络...

REM 启动透明代理服务
cd tools_docker_proxy
docker compose -f docker-compose.yml up -d
cd ..

REM 启动主服务
docker compose -f docker-compose.dev.yml up -d

echo.
echo 验证网络连接...
docker network inspect nofx-unified-network

echo.
echo 测试代理服务连通性...
timeout /t 5 /nobreak >nul
curl -s http://localhost:8081/health

echo.
echo ===============================================
echo    网络修复完成
echo ===============================================
echo.
echo 现在前端后端应该能够访问透明代理服务
pause