@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 更改生效验证脚本
echo ===============================================
echo.

echo 1. 检查服务运行状态...
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo.
echo 2. 验证网络连接...
echo 测试代理服务:
curl -s http://localhost:8081/health
echo.
echo 测试后端服务:
curl -s http://localhost:8888/api/health

echo.
echo 3. 检查关键日志...
echo.
echo --- 代理服务最近日志 ---
docker logs tools_docker_proxy --tail 5
echo.
echo --- 后端服务最近日志 ---
docker logs nofx-trading-dev --tail 5

echo.
echo 4. 验证环境变量...
docker exec nofx-trading-dev printenv | findstr TRANSPARENT_PROXY_URL

echo.
echo ===============================================
echo    验证完成
echo ===============================================
echo.
echo 如果所有测试都通过，说明更改已生效
echo 如果有问题，请检查:
echo 1. 网络配置是否正确
echo 2. 环境变量是否设置
echo 3. 防火墙是否阻止通信
echo.
pause