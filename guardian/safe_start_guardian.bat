@echo off
chcp 65001 >nul
echo 正在启动 AI 自动化守护程序 (安全模式)...
echo.

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
echo.

REM 检查Go是否安装
echo 检查Go环境...
go version
if errorlevel 1 (
    echo 错误: Go未安装或未在PATH中
    pause
    exit /b 1
)

REM 检查Chrome是否安装
echo 检查Chrome浏览器...
where chrome >nul 2>&1
if errorlevel 1 (
    echo Chrome未在PATH中找到，尝试常见安装路径...
    if exist "C:\Program Files\Google\Chrome\Application\chrome.exe" (
        set "GUARDIAN_CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe"
        echo 使用Chrome路径: %GUARDIAN_CHROME_PATH%
    ) else if exist "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" (
        set "GUARDIAN_CHROME_PATH=C:\Program Files (x86)\Google\Chrome\Application\chrome.exe"
        echo 使用Chrome路径: %GUARDIAN_CHROME_PATH%
    ) else (
        echo 警告: 未找到Chrome浏览器，程序可能无法启动
    )
) else (
    echo Chrome浏览器已找到
)

REM 检查依赖
echo 检查项目依赖...
go mod tidy
if errorlevel 1 (
    echo 警告: 依赖检查失败，但继续启动
)

echo.
echo 启动守护程序 (5秒后开始，按Ctrl+C取消)...
timeout /t 5 /nobreak >nul

REM 启动守护程序
go run main.go

echo.
echo 程序已退出，退出代码: %errorlevel%
pause