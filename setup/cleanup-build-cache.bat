@echo off
setlocal enabledelayedexpansion

echo ========================================
echo    NOFX 构建缓存清理脚本
echo ========================================
echo.

echo 正在清理Go构建缓存...
echo.

REM 清理Go构建缓存
echo [1/3] 清理Go构建缓存...
go clean -cache
if %errorlevel% equ 0 (
    echo Go构建缓存清理完成
) else (
    echo 警告: Go构建缓存清理可能未执行 (Go可能未安装)
)
echo.

REM 清理Go模块缓存
echo [2/3] 清理Go模块缓存...
go clean -modcache
if %errorlevel% equ 0 (
    echo Go模块缓存清理完成
) else (
    echo 警告: Go模块缓存清理可能未执行 (Go可能未安装)
)
echo.

REM 清理Docker构建缓存
echo [3/3] 清理Docker构建缓存...
docker builder prune -f
if %errorlevel% equ 0 (
    echo Docker构建缓存清理完成
) else (
    echo 警告: Docker构建缓存清理可能未执行 (Docker可能未安装或未运行)
)
echo.

echo ========================================
echo    构建缓存清理完成！
echo ========================================
echo.

echo 清理的缓存包括:
echo   - Go构建缓存 (%LOCALAPPDATA%\go-build)
echo   - Go模块缓存 (%GOPATH%/pkg/mod 或默认位置)
echo   - Docker构建缓存
echo.
echo 提示: 这些缓存文件通常可以安全删除，下次构建时会重新生成
echo.

pause