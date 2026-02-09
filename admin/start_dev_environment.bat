@echo off
echo.
echo ==========================================
echo    NOFX 管理员系统开发环境启动脚本
echo ==========================================
echo.

REM 检查是否安装了 Go
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未找到 Go，请先安装 Go 并将其添加到 PATH 环境变量中
    pause
    exit /b 1
)

REM 检查是否安装了 Node.js
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未找到 Node.js，请先安装 Node.js
    pause
    exit /b 1
)

REM 检查是否安装了 npm
where npm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ✓ npm 已安装
) else (
    echo 错误: 未找到 npm
    pause
    exit /b 1
)

echo.
echo 正在启动开发环境...
echo.

REM 启动后端服务 (在新窗口中)
start "NOFX Admin Backend" cmd /k "cd /d \"%~dp0\" && go run main.go"

timeout /t 3 /nobreak >nul

REM 启动前端开发服务器 (在新窗口中)
if exist web (
    start "NOFX Admin Frontend" cmd /k "cd /d \"%~dp0web\" && npm run dev"
    echo.
    echo ==========================================
    echo    开发环境已启动!
    echo ==========================================
    echo.
    echo 后端服务已在新窗口中启动
    echo 前端开发服务器已在新窗口中启动
    echo.
    echo 访问地址:
    echo - 管理员登录: http://localhost:3000/login
    echo - API 文档: http://localhost:9000/swagger (如果配置了)
    echo.
    echo 默认管理员账号:
    echo - 用户名: admin
    echo - 密码: admin123 (首次登录后请立即更改!)
    echo ==========================================
) else (
    echo.
    echo ==========================================
    echo    警告: 未找到前端目录 web
    echo    仅启动后端服务
    echo ==========================================
    echo.
    echo 后端服务已在新窗口中启动
    echo API 访问地址: http://localhost:9000/api/
    echo ==========================================
)

pause