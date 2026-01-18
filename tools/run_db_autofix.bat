@echo off
echo 正在运行数据库结构自动修复工具...
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
echo 开始修复数据库结构...
echo.

REM 运行SQL脚本来修复数据库结构
..\sqlite-tools\sqlite3.exe %DB_PATH% < autofix_db_structure.sql

echo.
echo 数据库结构自动修复完成！
echo.
echo 您现在可以重启后端系统以应用更改。
pause