@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX Docker 环境清理脚本
echo ========================================
echo.

echo 警告: 此操作将停止并删除所有相关的Docker容器、网络和镜像
echo 数据库文件 (data.db) 将被保留
echo.
set /p confirm=是否继续? (Y/N): 
if /i not "!confirm!"=="Y" (
    echo 操作已取消
    pause
    exit /b 0
)

echo.
echo [1/4] 停止并删除开发版容器...
docker-compose -f ../docker-compose.dev.watch.yml down -v
if %errorlevel% equ 0 (
    echo 开发版容器已清理
) else (
    echo 开发版容器清理可能失败或不存在
)

echo.
echo [2/4] 停止并删除稳定版容器...
docker-compose -f ../docker-compose.stable.yml down -v
if %errorlevel% equ 0 (
    echo 稳定版容器已清理
) else (
    echo 稳定版容器清理可能失败或不存在
)

echo.
echo [3/4] 清理Docker系统资源...
docker system prune -f
docker builder prune -f

echo.
echo [4/4] 检查残留容器...
for /f "tokens=*" %%i in ('docker ps -aq -f name=nofx') do (
    echo 停止并删除容器: %%i
    docker stop %%i >nul 2>&1
    docker rm %%i >nul 2>&1
)

echo.
echo ========================================
echo    环境清理完成！
echo ========================================
echo.
echo 提示:
echo   - 数据库文件 E:\AI\nofx_Dev\data\data.db 已保留
echo   - 如需彻底删除数据，请手动删除 data 目录
echo   - 现在可以重新运行安装脚本
echo ========================================
echo.

pause