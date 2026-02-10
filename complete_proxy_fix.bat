@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 透明代理完整修复脚本
echo ===============================================
echo.

echo 1. 停止当前所有服务...
docker compose -f docker-compose.dev.yml down 2>nul
docker compose -f tools_docker_proxy\docker-compose.yml down 2>nul

echo.
echo 2. 清理旧网络...
docker network rm nofx-network nofx-proxy-network 2>nul

echo.
echo 3. 创建统一网络...
docker network create nofx-unified-network

echo.
echo 4. 启动透明代理服务...
cd tools_docker_proxy
docker compose -f docker-compose.unified.yml up -d --build
cd ..

echo.
echo 5. 启动主服务...
docker compose -f docker-compose.dev.yml up -d --build

echo.
echo 6. 等待服务启动...
timeout /t 10 /nobreak >nul

echo.
echo 7. 验证服务状态...
echo.
echo --- 透明代理服务 ---
docker logs tools_docker_proxy --tail 10
echo.
echo --- 主服务 ---
docker logs nofx-trading-dev --tail 10

echo.
echo 8. 测试网络连通性...
echo 测试代理健康检查:
curl -s http://localhost:8081/health
echo.
echo 测试后端API:
curl -s http://localhost:8888/api/health

echo.
echo ===============================================
echo    修复完成！验证步骤：
echo ===============================================
echo 1. 访问前端: http://localhost:3300
echo 2. 检查代理日志: docker logs tools_docker_proxy -f
echo 3. 验证代理是否收到请求
echo.
echo 如果仍有问题，请检查:
echo - 后端是否正确配置了 TRANSPARENT_PROXY_URL 环境变量
echo - 请求是否包含 X-Custom-API-URL 头部
echo - 防火墙设置是否阻止了容器间通信
echo.
pause