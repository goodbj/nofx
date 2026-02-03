@echo off
chcp 65001 >nul
setlocal

echo ================================
echo 币安代理服务管理工具
echo ================================

:menu
echo.
echo 请选择操作:
echo 1. 查看服务状态
echo 2. 启动服务
echo 3. 停止服务
echo 4. 重启服务
echo 5. 查看日志
echo 6. 查看详细日志(实时)
echo 7. 删除容器
echo 8. 重新部署
echo 9. 退出

choice /c 123456789 /m "请输入选择(1-9)"

if errorlevel 9 goto :eof
if errorlevel 8 goto redeploy
if errorlevel 7 goto remove
if errorlevel 6 goto logs_live
if errorlevel 5 goto logs
if errorlevel 4 goto restart
if errorlevel 3 goto stop
if errorlevel 2 goto start
if errorlevel 1 goto status

:status
echo.
echo ================================
echo 服务状态检查
echo ================================
echo 检查Docker环境...
docker --version >nul 2>&1
if %errorLevel% neq 0 (
    echo ❌ Docker未安装或未运行
    goto menu
)

echo 检查容器状态...
docker ps -a | findstr binance-proxy-container >nul
if %errorLevel% equ 0 (
    echo.
    echo 容器信息:
    docker ps -a | findstr binance-proxy-container
    echo.
    
    docker ps | findstr binance-proxy-container >nul
    if %errorLevel% equ 0 (
        echo ✅ 服务正在运行
        echo 测试服务连接...
        curl -s http://localhost:8081/health >nul 2>&1
        if %errorLevel% equ 0 (
            echo ✅ 服务连接正常
            echo 服务地址: http://localhost:8081/health
        ) else (
            echo ⚠️  容器运行中但服务无响应
        )
    ) else (
        echo ❌ 服务已停止
    )
) else (
    echo ❌ 未找到代理服务容器
)
goto menu

:start
echo.
echo 启动币安代理服务...
docker start binance-proxy-container >nul 2>&1
if %errorLevel% equ 0 (
    echo ✅ 服务启动成功
    timeout /t 3 /nobreak >nul
    echo 测试服务连接...
    curl -s http://localhost:8081/health >nul 2>&1
    if %errorLevel% equ 0 (
        echo ✅ 服务连接正常
    ) else (
        echo ⚠️  服务启动中，请稍后检查
    )
) else (
    echo ❌ 启动失败，容器可能不存在
)
goto menu

:stop
echo.
echo 停止币安代理服务...
docker stop binance-proxy-container >nul 2>&1
if %errorLevel% equ 0 (
    echo ✅ 服务已停止
) else (
    echo ❌ 停止失败，服务可能未运行
)
goto menu

:restart
echo.
echo 重启币安代理服务...
docker restart binance-proxy-container >nul 2>&1
if %errorLevel% equ 0 (
    echo ✅ 服务重启成功
    timeout /t 3 /nobreak >nul
    echo 测试服务连接...
    curl -s http://localhost:8081/health >nul 2>&1
    if %errorLevel% equ 0 (
        echo ✅ 服务连接正常
    ) else (
        echo ⚠️  服务重启中，请稍后检查
    )
) else (
    echo ❌ 重启失败
)
goto menu

:logs
echo.
echo ================================
echo 服务日志 (最近50行)
echo ================================
docker logs --tail 50 binance-proxy-container
echo.
echo ================================
echo 按任意键返回菜单...
pause >nul
goto menu

:logs_live
echo.
echo ================================
echo 实时日志 (按 Ctrl+C 停止)
echo ================================
echo 启动实时日志监控...
docker logs -f binance-proxy-container
goto menu

:remove
echo.
echo ⚠️  警告: 此操作将删除容器和所有数据
choice /m "确定要删除容器吗"
if errorlevel 2 goto menu
if errorlevel 1 (
    echo 停止并删除容器...
    docker stop binance-proxy-container >nul 2>&1
    docker rm binance-proxy-container >nul 2>&1
    if %errorLevel% equ 0 (
        echo ✅ 容器已删除
    ) else (
        echo ❌ 删除失败
    )
)
goto menu

:redeploy
echo.
echo ⚠️  警告: 此操作将重新构建镜像
choice /m "确定要重新部署吗"
if errorlevel 2 goto menu
if errorlevel 1 (
    echo 停止现有服务...
    docker stop binance-proxy-container >nul 2>&1
    docker rm binance-proxy-container >nul 2>&1
    
    echo 重新构建镜像...
    cd api\binance_proxy
    if exist Dockerfile.simple (
        docker build -t binance-proxy-service -f Dockerfile.simple ../..
    ) else (
        docker build -t binance-proxy-service .
    )
    cd ..\..
    
    if %errorLevel% equ 0 (
        echo 启动新容器...
        docker run -d -p 8081:8081 --name binance-proxy-container binance-proxy-service
        if %errorLevel% equ 0 (
            echo ✅ 重新部署完成
        ) else (
            echo ❌ 容器启动失败
        )
    ) else (
        echo ❌ 镜像构建失败
    )
)
goto menu