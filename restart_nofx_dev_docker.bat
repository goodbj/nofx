@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 开发版 Docker 服务重启脚本
echo ===============================================
echo.
echo 正在重启服务...
echo.

:: 停止服务
echo 停止现有服务...
docker compose -f docker-compose.nofx-dev.yml down

:: 等待一段时间确保服务完全停止
timeout /t 5 /nobreak >nul

:: 启动服务
echo 启动服务...
docker compose -f docker-compose.nofx-dev.yml up -d

if %errorlevel% equ 0 (
    echo.
    echo ===============================================
    echo    NOFX 开发版服务已成功重启！
    echo ===============================================
    echo.
    echo 服务状态：
    docker compose -f docker-compose.nofx-dev.yml ps
    echo.
    echo 访问地址：
    echo   前端界面: http://localhost:3300
    echo   后端API: http://localhost:8888
    echo.
) else (
    echo.
    echo 错误：重启服务时出现问题
    echo 请检查错误信息
)

pause