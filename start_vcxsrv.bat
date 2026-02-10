@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    VcXsrv自动启动脚本
echo ===============================================
echo.

echo 正在启动XLaunch with 预配置...

REM 检查VcXsrv安装位置
set VCXSRV_PATH="C:\Program Files\VcXsrv\XLaunch.exe"
if not exist %VCXSRV_PATH% (
    set VCXSRV_PATH="C:\Program Files (x86)\VcXsrv\XLaunch.exe"
)

if not exist %VCXSRV_PATH% (
    echo ❌ 未找到VcXsrv安装，请确认是否已正确安装
    echo 请从 https://sourceforge.net/projects/vcxsrv/ 下载安装
    pause
    exit /b 1
)

echo ✅ 找到VcXsrv安装路径: %VCXSRV_PATH%

echo.
echo 启动XLaunch with 配置:
echo - Multiple windows 模式
echo - 显示编号: 0  
echo - 禁用访问控制
echo.

REM 启动XLaunch with 预配置参数
%VCXSRV_PATH% :0 -ac -clipboard -multiwindow

echo.
echo XLaunch已启动，请检查系统托盘中的X Server图标
echo 如果启动成功，系统托盘应该出现X Server图标

echo.
echo 等待5秒让X Server完全启动...
timeout /t 5 /nobreak >nul

echo.
echo 重新验证配置...
call verify_vcxsrv.bat

pause