@echo off
chcp 65001 >nul
echo 🚀 启动增强版AI自动化守护程序...
echo 📁 当前目录: %cd%
echo ⏰ 启动时间: %date% %time%
echo ===============================================

REM 设置环境变量
set GUARDIAN_USER_DATA_DIR=./chrome_profile

echo 🌐 Guardian将连接到NoFx后端: http://127.0.0.1:8888/api
echo 📁 使用Chrome用户数据目录: ./chrome_profile
echo.

echo 🔧 测试与NoFx后端的连接...
curl -s --connect-timeout 5 http://127.0.0.1:8888/api/health >nul 2>&1
if %errorlevel% == 0 (
    echo ✅ NoFx后端连接正常
) else (
    echo ⚠️ 无法连接到NoFx后端，请确保NoFx后端已在端口8888上运行
)

echo.
echo 🚀 启动AI自动化守护程序...
echo 💡 详细日志将显示Guardian的操作过程
echo 💡 按 Ctrl+C 可停止守护程序
echo ===============================================

REM 启动Go程序
go run main.go

pause