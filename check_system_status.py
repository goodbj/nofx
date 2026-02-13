import sqlite3

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== 系统状态检查 ===")

# 检查交易员配置
cursor.execute("""
    SELECT 
        id,
        name,
        scan_interval_minutes,
        is_running,
        last_run_time,
        next_run_time,
        error_count,
        last_error
    FROM traders 
    WHERE id = 'bac26dfa_guardian-ai_1770901210'
""")

trader_info = cursor.fetchone()
if trader_info:
    print(f"\n交易员配置:")
    print(f"  ID: {trader_info[0]}")
    print(f"  名称: {trader_info[1]}")
    print(f"  扫描间隔: {trader_info[2]} 分钟")
    print(f"  运行状态: {'运行中' if trader_info[3] else '停止'}")
    print(f"  上次运行: {trader_info[4] if trader_info[4] else '从未运行'}")
    print(f"  下次运行: {trader_info[5] if trader_info[5] else '未设置'}")
    print(f"  错误计数: {trader_info[6]}")
    print(f"  最后错误: {trader_info[7][:100] if trader_info[7] else '无'}")

# 检查系统时间
import datetime
print(f"\n当前系统时间: {datetime.datetime.now()}")

# 检查最近的错误记录
print(f"\n=== 最近错误记录 ===")
cursor.execute("""
    SELECT 
        timestamp,
        error_message
    FROM decision_records 
    WHERE trader_id = 'bac26dfa_guardian-ai_1770901210'
    AND success = 0
    ORDER BY timestamp DESC
    LIMIT 5
""")

error_records = cursor.fetchall()
for record in error_records:
    print(f"时间: {record[0][:19]}")
    print(f"错误: {record[1][:100] if record[1] else '无错误信息'}")
    print()

conn.close()