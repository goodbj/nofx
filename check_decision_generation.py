import sqlite3
from datetime import datetime, timedelta

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== 决策记录生成情况分析 ===")

# 获取最近24小时的决策记录
print("\n1. 最近24小时决策记录统计:")
cursor.execute("""
    SELECT 
        date(timestamp) as date,
        count(*) as record_count,
        min(timestamp) as first_record,
        max(timestamp) as last_record
    FROM decision_records 
    WHERE timestamp > datetime('now', '-1 day')
    GROUP BY date(timestamp)
    ORDER BY date DESC
""")

daily_stats = cursor.fetchall()
if daily_stats:
    for stat in daily_stats:
        print(f"日期: {stat[0]}, 记录数: {stat[1]}, 首条: {stat[2][:19]}, 末条: {stat[3][:19]}")
else:
    print("最近24小时无决策记录")

# 获取所有决策记录按时间排序
print("\n2. 所有决策记录时间分布:")
cursor.execute("""
    SELECT 
        trader_id,
        count(*) as total_records,
        min(timestamp) as first_record,
        max(timestamp) as last_record,
        max(cycle_number) as max_cycle
    FROM decision_records 
    GROUP BY trader_id
    ORDER BY max(timestamp) DESC
""")

trader_stats = cursor.fetchall()
for stat in trader_stats:
    print(f"交易员: {stat[0]}")
    print(f"  总记录数: {stat[1]}")
    print(f"  首条记录: {stat[2][:19]}")
    print(f"  最新记录: {stat[3][:19]}")
    print(f"  最大周期: {stat[4]}")
    print()

# 检查运行中的交易员状态
print("3. 运行中交易员状态:")
cursor.execute("""
    SELECT id, name, is_running, updated_at 
    FROM traders 
    WHERE is_running = 1
""")

running_traders = cursor.fetchall()
if running_traders:
    for trader in running_traders:
        print(f"交易员: {trader[1]} (ID: {trader[0]})")
        print(f"  运行状态: {'运行中' if trader[2] else '停止'}")
        print(f"  更新时间: {trader[3]}")
        print()
else:
    print("无运行中的交易员")

# 检查最近1小时是否有新记录生成
print("4. 最近1小时决策记录生成情况:")
one_hour_ago = datetime.now() - timedelta(hours=1)
cursor.execute("""
    SELECT count(*) as recent_count
    FROM decision_records 
    WHERE timestamp > ?
""", (one_hour_ago,))

recent_count = cursor.fetchone()[0]
print(f"最近1小时新增记录数: {recent_count}")

conn.close()