@echo off
REM *****************************************************
REM * Guardian自动化守护程序启动脚本(Docker环境兼容版) *
REM *****************************************************

echo 🚀 正在启动 AI 自动化守护程序...
echo.

REM 设置NoFx后端API端点(针对Docker环境的正确配置)
set NOFX_API_ENDPOINT=http://127.0.0.1:8888/api

REM 设置Guardian配置
set GUARDIAN_USER_DATA_DIR=./chrome_profile

REM 切换到Guardian目录
cd /d "e:\AI\nofx_Dev\guardian"

echo 🌐 Guardian将连接到NoFx后端: %NOFX_API_ENDPOINT%
echo 📁 使用Chrome用户数据目录: %GUARDIAN_USER_DATA_DIR%
echo.

REM 检查NoFx后端连接
echo 📡 测试与NoFx后端的连接...
ping -n 1 127.0.0.1 >nul
if errorlevel 1 (
  echo ⚠️  无法连接到本地主机，请检查网络
) else (
  echo ✅ 本地网络连接正常
)

REM 启动Guardian守护程序
echo 🚀 启动AI自动化守护程序...
go run main.go

if errorlevel 1 (
  echo.
  echo ❌ Guardian启动失败!
  echo 请检查:
  echo 1. NoFx后端是否在Docker中正常运行
  echo 2. 端口8888是否正确映射
  echo 3. 防火墙是否阻止连接
  pause
  exit /b 1
)

echo ✅ Guardian守护程序已启动
pause