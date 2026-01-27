@echo off
REM *****************************************************
REM * Guardian自动化守护程序启动脚本(调试模式)
REM *****************************************************

echo 正在启动 AI 自动化守护程序(调试模式)...
echo.

REM 设置调试环境变量
set NOFX_API_ENDPOINT=http://127.0.0.1:8888/api
set GUARDIAN_USER_DATA_DIR=./chrome_profile
set GUARDIAN_DEBUG=true

REM 切换到Guardian目录
cd /d "e:\AI\nofx_Dev\guardian"

echo Guardian将连接到NoFx后端: %NOFX_API_ENDPOINT%
echo 使用Chrome用户数据目录: %GUARDIAN_USER_DATA_DIR%
echo 启用调试模式: %GUARDIAN_DEBUG%
echo.

REM 启动Guardian守护程序(带调试输出)
echo 启动AI自动化守护程序(调试模式)...
go run -tags=debug main.go

if errorlevel 1 (
  echo.
  echo Guardian启动失败!
  echo 请检查:
  echo 1. NoFx后端是否在Docker中正常运行
  echo 2. 端口8888是否正确映射
  echo 3. 防火墙是否阻止连接
  pause
  exit /b 1
)

echo Guardian守护程序已启动
pause