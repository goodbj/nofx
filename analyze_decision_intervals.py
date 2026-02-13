import sqlite3
from datetime import datetime

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== 决策记录生成时间间隔分析 ===")

# 获取运行中交易员的决策记录时间戳
trader_id = "bac26dfa_guardian-ai_1770901210"
cursor.execute("""
    SELECT 
        cycle_number,
        timestamp,
        success,
        error_message
    FROM decision_records 
    WHERE trader_id = ?
    ORDER BY timestamp DESC
    LIMIT 20
""", (trader_id,))

records = cursor.fetchall()

print(f"\n交易员 {trader_id} 的最近决策记录:")
print("周期 | 时间戳 | 成功 | 错误信息")
print("-" * 80)

previous_time = None
intervals = []

for record in records:
    cycle = record[0]
    timestamp = record[1][:19]
    success = "✓" if record[2] else "✗"
    error_msg = record[3][:50] if record[3] else ""
    
    # 计算时间间隔
    current_time = datetime.strptime(timestamp, "%Y-%m-%d %H:%M:%S")
    if previous_time:
        interval = (previous_time - current_time).total_seconds() / 60  # 转换为分钟
        intervals.append(interval)
        interval_str = f"{interval:.1f}分钟"
    else:
        interval_str = "N/A"
    
    print(f"{cycle:4d} | {timestamp} | {success:2s} | {interval_str:8s} | {error_msg}")
    previous_time = current_time

# 计算平均间隔
if intervals:
    avg_interval = sum(intervals) / len(intervals)
    print(f"\n平均生成间隔: {avg_interval:.1f} 分钟")
    print(f"预期间隔(配置): 5 分钟")
    print(f"差异: {abs(avg_interval - 5):.1f} 分钟")

conn.close()