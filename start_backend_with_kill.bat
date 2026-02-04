@echo off
chcp 65001 >nul
setlocal

set SERVER_PORT=8888

echo 正在查找并终止占用端口 %SERVER_PORT% 的进程...
echo ============================

:: 查找占用端口的进程PID
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :%SERVER_PORT% ^| findstr LISTENING') do (
    set PID=%%a
)

:: 如果找到了PID，则终止进程
if defined PID (
    echo 找到占用端口 %SERVER_PORT% 的进程 PID: %PID%
    echo 正在终止进程...
    taskkill /pid %PID% /f
    if errorlevel 1 (
        echo 警告: 无法终止进程 PID: %PID%，可能权限不足或进程不存在
    ) else (
        echo 成功终止进程 PID: %PID%
    )
) else (
    echo 未找到占用端口 %SERVER_PORT% 的进程
)

echo.
echo 等待片刻让端口释放...
timeout /t 2 /nobreak >nul

echo.
echo 启动 NOFX 后端服务器...
echo ============================
echo 正在启动后端服务器，端口: %SERVER_PORT%
echo.

cd /d "e:\AI\nofx_Dev"

echo 启动命令: set API_SERVER_PORT=%SERVER_PORT% && set JWT_SECRET=dev-jwt-secret-change-in-production && set DATA_ENCRYPTION_KEY=ZGV2LWRhdGEtZW5jcnlwdGlvbi1rZXktZGV2LWRhdGE= && set OLLAMA_READ_TIMEOUT=270s && go run main.go

set API_SERVER_PORT=%SERVER_PORT% && set JWT_SECRET=dev-jwt-secret-change-in-production && set DATA_ENCRYPTION_KEY=ZGV2LWRhdGEtZW5jcnlwdGlvbi1rZXktZGV2LWRhdGE= && set OLLAMA_READ_TIMEOUT=270s && go run main.go

if errorlevel 1 (
    echo.
    echo 错误: 后端服务器启动失败
    pause
    exit /b 1
)

pause