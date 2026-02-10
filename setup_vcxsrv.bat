@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    VcXsrv一键配置脚本
echo ===============================================
echo.

echo 正在为您配置VcXsrv环境...

echo 1. 设置DISPLAY环境变量...
set DISPLAY=host.docker.internal:0.0
echo DISPLAY变量已设置为: %DISPLAY%

echo.
echo 2. 添加防火墙例外...
netsh advfirewall firewall add rule name="VcXsrv" dir=in action=allow program="C:\Program Files\VcXsrv\XWin.exe" >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 防火墙规则已添加
) else (
    echo ⚠️ 防火墙规则添加失败（可能已存在或需要管理员权限）
)

echo.
echo 3. 验证配置...
call verify_vcxsrv.bat

echo.
echo ===============================================
echo    配置完成！
echo ===============================================
echo.
echo 现在可以启动带显示的Docker服务了
echo 运行: deploy_with_display.bat
echo.
pause