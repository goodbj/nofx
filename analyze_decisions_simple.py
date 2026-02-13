import sqlite3
import json
from datetime import datetime

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

# 获取交易员列表
print("=== 交易员列表 ===")
cursor.execute("SELECT id, name, is_running FROM traders")
traders = cursor.fetchall()
for trader in traders:
    print(f"ID: {trader[0]}, Name: {trader[1]}, Running: {trader[2]}")

# 获取最新决策记录
print("\n=== 最新决策记录 ===")
for trader in traders:
    trader_id = trader[0]
    print(f"\n交易员: {trader[1]} (ID: {trader_id})")
    try:
        cursor.execute("""
            SELECT cycle_number, timestamp, success 
            FROM decision_records 
            WHERE trader_id = ? 
            ORDER BY timestamp DESC 
            LIMIT 5
        """, (trader_id,))
        
        records = cursor.fetchall()
        if records:
            print("Cycle | Timestamp | Success")
            print("-" * 40)
            for record in records:
                print(f"{record[0]:5d} | {record[1][:19]} | {bool(record[2])}")
        else:
            print("暂无决策记录")
    except Exception as e:
        print(f"查询错误: {e}")

conn.close()