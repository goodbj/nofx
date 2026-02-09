@echo off
echo.
echo ==========================================
echo    NOFX 管理员系统部署脚本
echo ==========================================
echo.

REM 检查是否安装了 Go
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未找到 Go，请先安装 Go 并将其添加到 PATH 环境变量中
    pause
    exit /b 1
)

REM 检查是否安装了 Node.js (用于前端)
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 警告: 未找到 Node.js，前端界面将无法运行
) else (
    echo ✓ Node.js 已安装
)

REM 检查是否安装了 npm
where npm >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 警告: 未找到 npm，前端界面将无法运行
) else (
    echo ✓ npm 已安装
)

echo.
echo 正在构建管理员系统...
echo.

REM 设置 Go 模块
echo 设置 Go 模块...
cd /d "%~dp0"
go mod tidy

if %ERRORLEVEL% neq 0 (
    echo 错误: Go 模块设置失败
    pause
    exit /b 1
)

REM 编译后端
echo.
echo 编译后端服务...
go build -o admin-server.exe .

if %ERRORLEVEL% neq 0 (
    echo 错误: 后端编译失败
    pause
    exit /b 1
)

echo ✓ 后端编译成功

REM 构建前端
if exist web (
    echo.
    echo 构建前端界面...
    cd web
    npm install
    npm run build
    cd ..
    echo ✓ 前端构建成功
) else (
    echo 警告: 未找到前端目录 web，跳过前端构建
)

echo.
echo ==========================================
echo    管理员系统部署完成!
echo ==========================================
echo.
echo 启动后端服务: .\admin-server.exe
echo 或者开发模式: go run main.go
echo.
echo 前端开发模式 (在 web 目录下): npm run dev
echo.
echo 默认访问地址:
echo - 后端 API: http://localhost:9000
echo - 前端界面: http://localhost:3000
echo.
echo 管理员登录地址: http://localhost:3000/login
echo ==========================================

pause