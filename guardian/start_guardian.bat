@echo off
chcp 65001 >nul
echo 正在启动 AI 自动化守护程序...
echo.

REM 设置NoFx后端API端点(Docker环境配置)
set NOFX_API_ENDPOINT=http://127.0.0.1:8888/api

REM 设置Guardian配置
set GUARDIAN_USER_DATA_DIR=./chrome_profile
set GUARDIAN_AI_PROVIDER=deepseek
set GUARDIAN_AI_ENDPOINT=https://chat.deepseek.com
set GUARDIAN_PAGE_URL=https://chat.deepseek.com

REM 可选：设置API密钥和认证令牌（如果需要）
REM set GUARDIAN_API_KEY=your_api_key_here
REM set GUARDIAN_AUTH_TOKEN=your_auth_token_here

REM 可选：设置Chrome路径（如果需要指定特定路径）
REM set GUARDIAN_CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe

echo Guardian将连接到NoFx后端: %NOFX_API_ENDPOINT%
echo 使用Chrome用户数据目录: %GUARDIAN_USER_DATA_DIR%
echo 使用AI提供商: %GUARDIAN_AI_PROVIDER%
echo.

REM 切换到Guardian目录
cd /d "e:\AI\nofx_Dev\guardian"

echo 启动守护程序...
go run main.go

pause