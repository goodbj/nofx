@echo off
chcp 65001 >nul
setlocal

echo 构建 NOFX 启动器可执行文件
echo.

REM 检查 Python 是否安装
python --version >nul 2>&1
if errorlevel 1 (
    echo 错误：未找到 Python，请先安装 Python
    pause
    exit /b 1
)

REM 检查是否安装了 PyInstaller
pip show PyInstaller >nul 2>&1
if errorlevel 1 (
    echo 未找到 PyInstaller，正在安装...
    pip install pyinstaller
    if errorlevel 1 (
        echo 安装 PyInstaller 失败
        pause
        exit /b 1
    )
)

echo.
echo 正在构建可执行文件...
pyinstaller --onefile --windowed --icon=NONE launcher_complete.py -n "NOFX_Launcher"

if not errorlevel 1 (
    echo.
    echo 构建成功！
    echo 可执行文件位于 dist 文件夹中
    echo 文件名: NOFX_Launcher.exe
    echo.
    echo 您可以直接运行 NOFX_Launcher.exe 来启动图形界面
) else (
    echo.
    echo 构建失败，请检查错误信息
)

pause