cd e:\AI\nofx_Dev\guardian
del go.bat
echo @echo off > go.bat
echo. >> go.bat
echo REM Guardian守护程序启动脚本 ^(Docker环境^) >> go.bat
echo REM 配置NoFx后端API端点 >> go.bat
echo set NOFX_API_ENDPOINT=http://127.0.0.1:8888/api >> go.bat
echo. >> go.bat
echo REM 配置Chrome用户数据目录^(保存登录状态^) >> go.bat
echo set GUARDIAN_USER_DATA_DIR=./chrome_profile >> go.bat
echo. >> go.bat
echo REM 启动Guardian守护程序 >> go.bat
echo echo 🚀 Starting AI Automation Guardian for Docker Environment... >> go.bat
echo echo. >> go.bat
echo go run main.go >> go.bat
echo. >> go.bat
echo if errorlevel 1 ^( >> go.bat
echo   echo ❌ Guardian启动失败，请检查： >> go.bat
echo   echo 1. NoFx后端服务是否正在运行 >> go.bat
echo   echo 2. 端口8888是否正确映射 >> go.bat
echo   echo 3. 网络连接是否正常 >> go.bat
echo ^) >> go.bat