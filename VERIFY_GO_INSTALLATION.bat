@echo off
chcp 65001 >nul
echo.
echo ========================================
echo🧪 GO环境验证脚本
echo ========================================
echo.

echo正在验证 GO环境...
echo.

REM 检查 go命
echo [1/4]检查 go命令...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 未找到 go命
    echo    请确保：
    echo    1. GO已正确安装
    echo    2.命行窗口已重启
    echo    3. PATH 环境变量包含 GO 的 bin目录
    echo.
    echo 当前 PATH 中的 GO相关路径：
    echo %PATH% | findstr /I "go"
    goto end
) else (
    echo✅ go命令可用
    go version
)
echo.

REM 检查 GOROOT
echo [2/4]检查 GOROOT环境变量...
go env GOROOT >nul 2>&1
if %errorlevel% neq 0 (
    echo⚠️  无法获取 GOROOT 信息
) else (
    echo✅ GOROOT配置正确
    echo GOROOT: %GOROOT%
    go env GOROOT
)
echo.

REM 检查 GOPATH
echo [3/4]检查 GOPATH环境变量...
go env GOPATH >nul 2>&1
if %errorlevel% neq 0 (
    echo⚠️  无法获取 GOPATH 信息
) else (
    echo ✅ GOPATH配置正确
    echo GOPATH: %GOPATH%
    go env GOPATH
)
echo.

REM测试基本 GO命
echo [4/4]测试基本 GO功能...
echo 创建临时测试文件...
echo package main > test.go
echo func main() { println("Hello, GO!") } >> test.go

go run test.go >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ GO运行测试失败
    echo   可能的问题：
    echo    1. GO安装不完整
    echo    2.权问题
    echo    3. 系统环境问题
) else (
    echo✅ GO运行测试通过
    echo GO环境配置正确！
)
echo.

REM 清理测试文件
del test.go >nul 2>&1

echo ========================================
echo验证结果:
echo ========================================
echo 如果所有检查都通过，您现在可以运行：
echo   - NOFX后端: cd E:\AI\nofx_dev && go run main.go
echo   - 透明代理: cd E:\AI\nofx_dev\tools_docker_proxy && go run main.go
echo.
echo 如果遇到问题，请检查：
echo   1. 是否已重启命令行窗口
echo   2. 环境变量是否正确配置
echo   3.墙是否阻止了 GO
echo ========================================
:end
echo.
echo按任意键退出...
pause >nul