@echo off
chcp 65001 >nul
echo AI自动化守护程序 - 带延迟的重启模式
echo.

set RESTART_COUNT=0
set MAX_RESTARTS=10
set NORMAL_DELAY=5
set ERROR_DELAY=10

REM 设置环境变量
set NOFX_API_ENDPOINT=http://127.0.0.1:8888/api
set GUARDIAN_USER_DATA_DIR=./chrome_profile_safe
set GUARDIAN_AI_PROVIDER=deepseek
set GUARDIAN_AI_ENDPOINT=https://chat.deepseek.com
set GUARDIAN_PAGE_URL=https://chat.deepseek.com
set GUARDIAN_LISTEN_INTERVAL=5000

echo Guardian将连接到NoFx后端: %NOFX_API_ENDPOINT%
echo 使用Chrome用户数据目录: %GUARDIAN_USER_DATA_DIR%
echo 使用AI提供商: %GUARDIAN_AI_PROVIDER%
echo 最大重启次数: %MAX_RESTARTS%
echo 正常重启延迟: %NORMAL_DELAY%秒, 错误重启延迟: %ERROR_DELAY%秒
echo.

:START
set /a RESTART_COUNT+=1
echo [%date% %time%] 第 %RESTART_COUNT% 次启动守护程序 (最大 %MAX_RESTARTS% 次)...

REM 检查Chrome
where chrome >nul 2>&1
if errorlevel 1 (
    if exist "C:\Program Files\Google\Chrome\Application\chrome.exe" (
        set "GUARDIAN_CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe"
    ) else if exist "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" (
        set "GUARDIAN_CHROME_PATH=C:\Program Files (x86)\Google\Chrome\Application\chrome.exe"
    )
)

REM 检查Go
go version >nul 2>&1
if errorlevel 1 (
    echo [%date% %time%] 错误: Go未安装或未在PATH中
    pause
    exit /b 1
)

REM 检查依赖
go mod tidy >nul 2>&1
if errorlevel 1 (
    echo [%date% %time%] 警告: 依赖检查失败
)

REM 启动守护程序
echo [%date% %time%] 开始运行守护程序...
go run main.go
set EXIT_CODE=%errorlevel%

echo [%date% %time%] 守护程序已退出，退出代码: %EXIT_CODE%

REM 检查是否达到最大重启次数
if %RESTART_COUNT% GEQ %MAX_RESTARTS% (
    echo [%date% %time%] 达到最大重启次数 %MAX_RESTARTS%，退出...
    pause
    exit /b 1
)

REM 根据退出代码决定延迟时间
if %EXIT_CODE% equ 0 (
    set DELAY=%NORMAL_DELAY%
    echo [%date% %time%] 正常退出，%DELAY%秒后重启...
) else (
    set DELAY=%ERROR_DELAY%
    echo [%date% %time%] 异常退出，%DELAY%秒后重启...
)

REM 等待延迟
for /L %%i in (1,1,%DELAY%) do (
    timeout /t 1 /nobreak >nul
    echo - 等待 %%i/%DELAY% 秒... (按Ctrl+C取消)
)

goto START