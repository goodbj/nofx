@echo off
REM 持仓状态自动检查服务启动脚本

echo === 持仓状态自动检查服务 ===
echo 正在启动后台服务...

REM 编译服务
go build auto_position_fix_service.go
if %errorlevel% equ 0 (
    echo 服务编译成功！
    echo 启动自动检查服务...
    echo 服务将每小时自动检查一次持仓状态
    echo 按 Ctrl+C 停止服务
    echo.
    
    REM 启动服务
    auto_position_fix_service.exe
) else (
    echo 服务编译失败
    pause
    exit /b 1
)

pause