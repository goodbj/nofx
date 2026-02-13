import sqlite3
import datetime

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
        updated_at
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
    print(f"  更新时间: {trader_info[4][:19] if trader_info[4] else '无'}")

print(f"\n当前系统时间: {datetime.datetime.now()}")

# 检查最近的决策记录
print(f"\n=== 最近决策记录时间线 ===")
cursor.execute("""
    SELECT 
        cycle_number,
        timestamp,
        success
    FROM decision_records 
    WHERE trader_id = 'bac26dfa_guardian-ai_1770901210'
    ORDER BY timestamp DESC
    LIMIT 10
""")

recent_records = cursor.fetchall()
for record in recent_records:
    print(f"周期 {record[0]:2d} | {record[1][:19]} | {'成功' if record[2] else '失败'}")

# 计算时间间隔
if len(recent_records) > 1:
    print(f"\n=== 时间间隔分析 ===")
    intervals = []
    for i in range(len(recent_records) - 1):
        current_time = datetime.datetime.strptime(recent_records[i][1][:19], "%Y-%m-%d %H:%M:%S")
        next_time = datetime.datetime.strptime(recent_records[i+1][1][:19], "%Y-%m-%d %H:%M:%S")
        interval = (current_time - next_time).total_seconds() / 60
        intervals.append(interval)
        print(f"周期 {recent_records[i][0]} → {recent_records[i+1][0]}: {interval:.1f} 分钟")
    
    if intervals:
        avg_interval = sum(intervals) / len(intervals)
        print(f"\n平均间隔: {avg_interval:.1f} 分钟")
        print(f"配置间隔: {trader_info[2]} 分钟")
        print(f"差异: {abs(avg_interval - trader_info[2]):.1f} 分钟")

conn.close()