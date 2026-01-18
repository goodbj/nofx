@echo off
echo 正在更新交易员状态...

REM 设置数据库路径
set DB_PATH=data\data.db

REM 检查数据库文件是否存在
if not exist "%DB_PATH%" (
    echo 错误: 数据库文件 %DB_PATH% 不存在
    pause
    exit /b 1
)

echo 更新交易员状态: 虚拟盘交易员设为非运行状态，实盘交易员保持运行状态
.\sqlite-tools\sqlite3.exe %DB_PATH% "UPDATE traders SET is_running = 0 WHERE name LIKE '%虚拟盘%';"

echo 更新剩余交易员状态完成

echo.
echo 当前交易员状态:
.\sqlite-tools\sqlite3.exe %DB_PATH% "SELECT id, name, is_running FROM traders ORDER BY name;"

echo.
echo 操作完成！
pause