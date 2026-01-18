@echo off
echo 正在清除AI模型表数据...
echo.

REM 检查是否存在sqlite3命令
sqlite3 -version >nul 2>&1
if errorlevel 1 (
    echo 错误: 未找到sqlite3命令
    echo 请先安装SQLite3
    pause
    exit /b 1
)

REM 检查数据库文件是否存在
if not exist "data\data.db" (
    echo 错误: 数据库文件 data\data.db 不存在
    pause
    exit /b 1
)

echo 备份当前AI模型数据...
copy "data\data.db" "data\data.db.backup_%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%.bak"

echo.
echo 当前AI模型表状态:
sqlite3 "data\data.db" "SELECT 'AI Models Count: ' || COUNT(*) FROM ai_models;"

echo.
echo 执行AI模型数据清除...
sqlite3 "data\data.db" < scripts\clear_ai_models.sql

echo.
echo 清除完成后的AI模型表状态:
sqlite3 "data\data.db" "SELECT 'AI Models Count: ' || COUNT(*) FROM ai_models;"

echo.
echo 操作完成！
pause