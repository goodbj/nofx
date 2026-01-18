@echo off
echo 正在重启NOFX后端系统...
echo.

REM 关闭可能正在运行的后端进程
echo 正在关闭可能正在运行的后端进程...
taskkill /f /im nofx.exe 2>nul
taskkill /f /im main.exe 2>nul

timeout /t 3 /nobreak >nul

REM 启动后端系统
echo 正在启动NOFX后端系统...
cd /d "e:\AI\nofx_Dev"

REM 检查可执行文件是否存在
if exist "nofx.exe" (
    start /min nofx.exe
    echo NOFX后端系统已在后台启动
) else if exist "main.exe" (
    start /min main.exe
    echo NOFX后端系统已在后台启动
) else (
    echo.
    echo 未找到可执行文件，正在尝试使用Go运行...
    if exist "go.mod" (
        start /min cmd /c "go run main.go"
        echo NOFX后端系统已在后台启动
    ) else (
        echo 错误: 未找到可执行文件或Go模块
    )
)

echo.
echo 系统将在后台运行，请稍等几分钟让系统完全启动
echo 然后您可以打开浏览器访问 http://localhost:8080 查看前端界面
echo.

pause