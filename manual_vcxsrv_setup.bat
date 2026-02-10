@echo off
chcp 65001 >nul
echo.
echo ===============================================
echo    简化版VcXsrv启动脚本
echo ===============================================
echo.

echo 请按照以下步骤手动操作：

echo.
echo 步骤1：启动XLaunch
echo ==================
echo 1. 按 Win + R 键
echo 2. 输入: XLaunch
echo 3. 按回车键

echo.
echo 步骤2：配置XLaunch（非常重要！）
echo =========================
echo 请按以下顺序选择配置：
echo.
echo 第1屏：选择启动方式
echo → 选择 "Multiple windows" （第一个选项）
echo → 点击 "Next"
echo.
echo 第2屏：选择显示编号  
echo → 选择显示编号 "0"
echo → 点击 "Next" 
echo.
echo 第3屏：客户端启动
echo → 选择 "Start no client"
echo → 点击 "Next"
echo.
echo 第4屏：额外参数
echo → 勾选 "Disable access control"
echo → 点击 "Next"
echo.
echo 第5屏：完成
echo → 点击 "Finish"

echo.
echo 步骤3：验证启动
echo ==============
echo 启动后请检查：
echo ✅ 系统托盘中是否出现X Server图标
echo ✅ 任务管理器中是否有XWin.exe进程

echo.
echo 步骤4：设置环境变量
echo =================
echo 打开新的命令提示符窗口，运行：
echo set DISPLAY=host.docker.internal:0.0

echo.
echo 步骤5：验证配置
echo ==============
echo 运行: verify_vcxsrv.bat

echo.
echo ===============================================
echo    重要提示
echo ===============================================
echo.
echo ❗ 如果配置正确，您应该：
echo 1. 看到系统托盘中的X Server图标
echo 2. verify_vcxsrv.bat显示所有检查通过
echo 3. 能够启动带显示的Docker服务

echo.
echo 如果遇到问题：
echo 1. 重新启动XLaunch
echo 2. 检查防火墙设置
echo 3. 确保以管理员身份运行
echo.
pause