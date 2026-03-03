@echo off
chcp 65001 >nul
echo.
echo ========================================
echo 🛠️  NOFX 开发环境检查工具
echo ========================================
echo.

echo 检查必需的开发工具...
echo.

REM 检查 Go
echo [1/4] 检查 Go 环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go 未安装或未添加到PATH
    echo    请访问 https://golang.org/dl/ 下载并安装Go
    echo    安装后请确保将Go的bin目录添加到系统PATH
    set GO_INSTALLED=0
) else (
    echo ✅ Go 已安装
    go version
    set GO_INSTALLED=1
)
echo.

REM 检查 Docker
echo [2/4] 检查 Docker 环境...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未安装
    echo    请访问 https://www.docker.com/products/docker-desktop 下载并安装Docker Desktop
    set DOCKER_INSTALLED=0
) else (
    echo ✅ Docker 已安装
    docker --version
    set DOCKER_INSTALLED=1
)
echo.

REM 检查 Node.js
echo [3/4] 检查 Node.js 环境...
node --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  Node.js 未安装（前端开发可能需要）
    echo    请访问 https://nodejs.org/ 下载并安装Node.js
    set NODE_INSTALLED=0
) else (
    echo ✅ Node.js 已安装
    node --version
    set NODE_INSTALLED=1
)
echo.

REM 检查 Git
echo [4/4] 检查 Git 环境...
git --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  Git 未安装（版本控制需要）
    echo    请访问 https://git-scm.com/ 下载并安装Git
    set GIT_INSTALLED=0
) else (
    echo ✅ Git 已安装
    git --version
    set GIT_INSTALLED=1
)
echo.

echo ========================================
echo 环境检查结果:
echo ========================================
if defined GO_INSTALLED (
    if "%GO_INSTALLED%"=="1" (
        echo ✅ Go: 已安装并配置
    ) else (
        echo ❌ Go: 未安装或配置错误
    )
)
if defined DOCKER_INSTALLED (
    if "%DOCKER_INSTALLED%"=="1" (
        echo ✅ Docker: 已安装并配置
    ) else (
        echo ❌ Docker: 未安装或配置错误
    )
)
if defined NODE_INSTALLED (
    if "%NODE_INSTALLED%"=="1" (
        echo ✅ Node.js: 已安装
    ) else (
        echo ⚠️  Node.js: 未安装
    )
)
if defined GIT_INSTALLED (
    if "%GIT_INSTALLED%"=="1" (
        echo ✅ Git: 已安装
    ) else (
        echo ⚠️  Git: 未安装
    )
)
echo.

echo ========================================
echo 建议操作:
echo ========================================
echo 如果 Go 未安装:
echo   1. 访问 https://golang.org/dl/
echo   2. 下载并安装最新版本的Go
echo   3. 安装完成后重启命令行窗口
echo.
echo 如果 Docker 未安装:
echo   1. 访问 https://www.docker.com/products/docker-desktop
echo   2. 下载并安装Docker Desktop
echo   3. 启动Docker Desktop应用程序
echo.
echo 安装完成后，您可以运行:
echo   - 后端: go run main.go
echo   - 透明代理: cd tools_docker_proxy && .\QUICK_START.bat
echo.

echo 按任意键退出...
pause >nul