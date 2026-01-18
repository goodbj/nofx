@echo off
echo 正在清空AI相关数据表...
echo.

REM 设置数据库路径
set DB_PATH=..\data\data.db

REM 检查数据库文件是否存在
if not exist "%DB_PATH%" (
    echo 警告: 根目录下的数据库文件 %DB_PATH% 不存在
    set DB_PATH=data\data.db
    if not exist "%DB_PATH%" (
        echo 错误: 数据库文件 %DB_PATH% 不存在
        pause
        exit /b 1
    ) else (
        echo 找到数据库文件: %DB_PATH%
    )
) else (
    echo 找到数据库文件: %DB_PATH%
)

echo.
echo 开始清空AI相关数据表...
echo.

REM 运行SQL脚本来清空AI相关数据表
echo 当前AI相关数据表记录数统计:
..\sqlite-tools\sqlite3.exe %DB_PATH% "SELECT 'decision_records:'||COUNT(*), 'trader_positions:'||COUNT(*), 'trader_orders:'||COUNT(*), 'trader_equity_snapshots:'||COUNT(*) FROM decision_records, trader_positions, trader_orders, trader_equity_snapshots;"

echo.
type clear_ai_related_tables.sql | ..\sqlite-tools\sqlite3.exe %DB_PATH%

echo.
echo AI相关数据表已清空完成！
echo.
echo 接下来请运行 restart_backend.bat 重启后端系统。
pause