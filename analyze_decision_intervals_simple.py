import sqlite3
from datetime import datetime

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== 决策记录生成时间间隔分析 ===")

# 获取运行中交易员的决策记录时间戳（避免error_message列）
trader_id = "bac26dfa_guardian-ai_1770901210"
cursor.execute("""
    SELECT 
        cycle_number,
        timestamp,
        success
    FROM decision_records 
    WHERE trader_id = ?
    ORDER BY timestamp DESC
    LIMIT 20
""", (trader_id,))

records = cursor.fetchall()

print(f"\n交易员 {trader_id} 的最近决策记录:")
print("周期 | 时间戳 | 成功")
print("-" * 50)

previous_time = None
intervals = []

for record in records:
    cycle = record[0]
    timestamp = record[1][:19]
    success = "✓" if record[2] else "✗"
    
    # 计算时间间隔
    current_time = datetime.strptime(timestamp, "%Y-%m-%d %H:%M:%S")
    if previous_time:
        interval = (previous_time - current_time).total_seconds() / 60  # 转换为分钟
        intervals.append(interval)
        interval_str = f"{interval:.1f}分钟"
    else:
        interval_str = "N/A"
    
    print(f"{cycle:4d} | {timestamp} | {success:2s} | {interval_str}")
    previous_time = current_time

# 计算平均间隔
if intervals:
    avg_interval = sum(intervals) / len(intervals)
    print(f"\n平均生成间隔: {avg_interval:.1f} 分钟")
    print(f"预期间隔(配置): 5 分钟")
    print(f"差异: {abs(avg_interval - 5):.1f} 分钟")

# 检查最近1小时的记录
print(f"\n=== 最近1小时记录统计 ===")
cursor.execute("""
    SELECT 
        count(*) as count,
        max(timestamp) as last_record
    FROM decision_records 
    WHERE trader_id = ?
    AND timestamp > datetime('now', '-1 hour')
""", (trader_id,))

recent_stats = cursor.fetchone()
print(f"最近1小时记录数: {recent_stats[0]}")
print(f"最新记录时间: {recent_stats[1][:19] if recent_stats[1] else '无'}")

conn.close()