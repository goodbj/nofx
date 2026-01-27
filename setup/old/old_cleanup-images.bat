@echo off
setlocal enabledelayedexpansion

echo ============================================
echo NOFX Docker 镜像清理脚本
echo ============================================
echo.
echo 正在列出所有镜像...
echo.

REM 显示所有镜像
docker images

echo.
echo ============================================
echo 保留的镜像（不会被删除）:
echo - nofx_dev-nofx-dev-watch (开发版后端)
echo - nofx_dev-nofx-frontend-dev-watch (开发版前端) 
echo - ghcr.io/nofxaios/nofx/nofx-backend:stable (稳定版后端)
echo - ghcr.io/nofxaios/nofx/nofx-frontend:stable (稳定版前端)
echo - golang:1.25-alpine (构建依赖)
echo - node:20-alpine (构建依赖)
echo - alpine:latest (基础镜像)
echo - 所有带有 'docker.m.daocloud.io' 前缀的镜像 (国内镜像源)
echo ============================================
echo.

REM 提示用户确认
echo 警告: 此操作将删除除上述镜像之外的所有 nofx 相关镜像
set /p confirm="是否继续? (y/N): "
if /i not "!confirm!"=="y" (
    echo 操作已取消
    pause
    exit /b 0
)

echo.
echo 正在查找需要清理的镜像...
echo.

REM 查找并删除不需要的镜像
echo 查找临时构建镜像...
for /f "skip=1" %%i in ('docker images --format "{{.Repository}} {{.Tag}} {{.ID}}" ^| findstr -v "nofx_dev-nofx-dev-watch\|nofx_dev-nofx-frontend-dev-watch\|ghcr.io/nofxaios/nofx/nofx-backend\|ghcr.io/nofxaios/nofx/nofx-frontend\|golang\|node\|alpine\|docker.m.daocloud.io"') do (
    set "image_line=%%i %%j %%k"
    for /f "tokens=3" %%a in ("!image_line!") do (
        echo 删除镜像: %%a
        docker rmi -f %%a >nul 2>&1
    )
)

echo.
echo 镜像清理完成！
echo.

REM 显示清理后的镜像列表
echo 清理后的镜像列表:
docker images

echo.
echo ============================================
echo 镜像清理完成
echo ============================================

pause