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

# 查看traders表结构
print("\n=== traders 表结构 ===")
try:
    cursor.execute("PRAGMA table_info(traders)")
    columns = cursor.fetchall()
    for col in columns:
        print(f"{col[1]} ({col[2]})")
except:
    print("traders 表不存在")

# 查看traders表数据
print("\n=== traders 表数据 ===")
try:
    cursor.execute("SELECT id, name, is_running FROM traders")
    results = cursor.fetchall()
    print("ID | Name | Is Running")
    print("-" * 30)
    for r in results:
        print(f"{r[0]} | {r[1]} | {r[2]}")
except Exception as e:
    print(f"查询失败: {e}")

# 查看decision_records表结构
print("\n=== decision_records 表结构 ===")
try:
    cursor.execute("PRAGMA table_info(decision_records)")
    columns = cursor.fetchall()
    for col in columns:
        print(f"{col[1]} ({col[2]})")
except:
    print("decision_records 表不存在")

conn.close()