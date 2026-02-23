@echo off
REM 持仓状态一致性检查一键执行脚本

echo === 持仓状态一致性检查工具 ===
echo 正在编译检查工具...

REM 进入工具目录
cd /d "%~dp0\position_fix"

REM 编译工具
go build position_consistency_check.go
if %errorlevel% equ 0 (
    echo 编译成功！
    echo 开始执行检查...
    echo.
    
    REM 执行检查
    position_consistency_check.exe
    
    echo.
    echo 检查完成！
) else (
    echo 编译失败，请检查代码或依赖
    pause
    exit /b 1
)

pause