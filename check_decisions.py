import sqlite3

# 连接数据库
conn = sqlite3.connect('nofx.db')
cursor = conn.cursor()

# 查看所有表
print("=== 数据库中的表 ===")
cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
tables = cursor.fetchall()
for table in tables:
    print(table[0])

# 查看决策记录表结构
print("\n=== decision_records 表结构 ===")
try:
    cursor.execute("PRAGMA table_info(decision_records)")
    columns = cursor.fetchall()
    for col in columns:
        print(f"{col[1]} ({col[2]})")
except:
    print("decision_records 表不存在")

# 查看最近的决策记录
print("\n=== 最近的决策记录 ===")
try:
    cursor.execute("SELECT trader_id, cycle_number, timestamp, success FROM decision_records ORDER BY timestamp DESC LIMIT 10")
    results = cursor.fetchall()
    print("Trader ID | Cycle | Timestamp | Success")
    print("-" * 50)
    for r in results:
        print(f"{r[0]} | {r[1]} | {r[2]} | {r[3]}")
except Exception as e:
    print(f"查询失败: {e}")

conn.close()