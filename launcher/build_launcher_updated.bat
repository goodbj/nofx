@echo off
chcp 65001 >nul
setlocal

echo NOFX 启动器构建工具
echo.

echo 请选择要构建的版本：
echo 1. 标准版启动器 (基础功能)
echo 2. 高级版启动器 (带日志过滤功能)
echo 3. 构建全部版本
echo.

set /p choice="请输入选择 (1-3): "

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
if "%choice%"=="1" (
    echo 正在构建标准版启动器...
    pyinstaller --onefile --windowed --icon=NONE launcher_complete.py -n "NOFX_Launcher_Standard"
    if not errorlevel 1 (
        echo.
        echo 标准版启动器构建成功！
        echo 可执行文件位于 dist 文件夹中
        echo 文件名: NOFX_Launcher_Standard.exe
    ) else (
        echo 标准版启动器构建失败
    )
) else if "%choice%"=="2" (
    echo 正在构建高级版启动器...
    pyinstaller --onefile --windowed --icon=NONE launcher_advanced_fixed.py -n "NOFX_Launcher_Advanced"
    if not errorlevel 1 (
        echo.
        echo 高级版启动器构建成功！
        echo 可执行文件位于 dist 文件夹中
        echo 文件名: NOFX_Launcher_Advanced.exe
    ) else (
        echo 高级版启动器构建失败
    )
) else if "%choice%"=="3" (
    echo 正在构建标准版启动器...
    pyinstaller --onefile --windowed --icon=NONE launcher_complete.py -n "NOFX_Launcher_Standard"
    if not errorlevel 1 (
        echo 标准版启动器构建成功
    ) else (
        echo 标准版启动器构建失败
    )
    
    echo.
    echo 正在构建高级版启动器...
    pyinstaller --onefile --windowed --icon=NONE launcher_advanced_fixed.py -n "NOFX_Launcher_Advanced"
    if not errorlevel 1 (
        echo 高级版启动器构建成功
    ) else (
        echo 高级版启动器构建失败
    )
    
    echo.
    echo 两个版本都已尝试构建完成
) else (
    echo 无效选择，退出
    pause
    exit /b 1
)

echo.
echo 您可以直接运行生成的 .exe 文件来启动图形界面
pause