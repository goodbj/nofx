@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    VcXsrv配置验证脚本
echo ===============================================
echo.

echo 1. 检查X Server进程...
tasklist | findstr XWin.exe >nul
if %errorlevel% equ 0 (
    echo ✅ X Server正在运行
) else (
    echo ❌ X Server未运行，请先启动XLaunch
    goto :end
)

echo.
echo 2. 检查DISPLAY环境变量...
echo 当前DISPLAY值: %DISPLAY%
if "%DISPLAY%"=="host.docker.internal:0.0" (
    echo ✅ DISPLAY环境变量设置正确
) else (
    echo ⚠️  DISPLAY环境变量可能需要设置
    echo 请运行: set DISPLAY=host.docker.internal:0.0
)

echo.
echo 3. 测试Docker容器连接...
docker run --rm -e DISPLAY=host.docker.internal:0.0 alpine sh -c "apk add --no-cache xeyes && xeyes" 2>nul
if %errorlevel% equ 0 (
    echo ✅ Docker容器可以连接到X Server
    echo 如果配置正确，您应该看到一个眼睛跟随鼠标移动的窗口
) else (
    echo ❌ Docker容器无法连接到X Server
    echo 可能的原因：
    echo 1. X Server未正确启动
    echo 2. 防火墙阻止了连接
    echo 3. DISPLAY变量设置不正确
)

echo.
echo 4. 防火墙检查...
netsh advfirewall firewall show rule name="VcXsrv" >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 已找到VcXsrv防火墙规则
) else (
    echo ⚠️ 建议为VcXsrv添加防火墙例外
    echo 运行以下命令（需要管理员权限）：
    echo netsh advfirewall firewall add rule name="VcXsrv" dir=in action=allow program="C:\Program Files\VcXsrv\XWin.exe"
)

echo.
echo 5. 常见问题排查...
echo.
echo 如果仍然无法显示：
echo 1. 重启XLaunch，确保选择了正确的配置
echo 2. 检查Windows防火墙设置
echo 3. 尝试关闭Windows Defender实时保护临时测试
echo 4. 确保Docker Desktop正在运行
echo 5. 检查是否有其他X Server程序冲突

:end
echo.
echo ===============================================
echo    验证完成
echo ===============================================
echo.
echo 如果所有检查都通过，现在可以启动带显示的Docker服务了
echo 运行: deploy_with_display.bat
echo.
pause