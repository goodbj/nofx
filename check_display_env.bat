@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    NOFX 浏览器显示环境检查脚本
echo ===============================================
echo.

echo 1. 检查Docker环境...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker未安装或未运行
    goto :end
)
echo ✅ Docker环境正常

echo.
echo 2. 检查显示服务器...
echo Windows环境下，浏览器显示通常需要以下配置之一：
echo - X11转发服务器 (如Xming, VcXsrv)
echo - 使用Windows子系统for Linux (WSL2) 
echo - 本地Chrome安装

echo.
echo 3. 推荐的Windows显示解决方案：

echo.
echo 方案A: 使用VcXsrv (推荐)
echo 1. 下载并安装VcXsrv: https://sourceforge.net/projects/vcxsrv/
echo 2. 启动XLaunch，选择"Multiple windows"
echo 3. 配置显示编号为0
echo 4. 启用"Disable access control"
echo 5. 启动后设置环境变量:
echo    set DISPLAY=host.docker.internal:0.0

echo.
echo 方案B: 使用WSL2 + WSLg
echo 1. 确保已安装WSL2
echo 2. 在WSL2中运行Docker
echo 3. 浏览器窗口将自动显示

echo.
echo 4. 当前容器显示配置检查...
docker exec nofx-dev-backend-display printenv | findstr -i "display\|chrome"
if %errorlevel% neq 0 (
    echo 未找到相关显示环境变量
)

echo.
echo 5. 测试显示功能...
echo 尝试在后端触发一个简单的浏览器操作来测试显示...

:end
echo.
echo ===============================================
echo    检查完成
echo ===============================================
pause