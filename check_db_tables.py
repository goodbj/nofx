import sqlite3

# 连接数据库
conn = sqlite3.connect('nofx.db')
cursor = conn.cursor()

# 获取所有表名
cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
tables = cursor.fetchall()

print("📋 Tables in database:")
for table in tables:
    print(f"  - {table[0]}")

# 检查是否有包含决策信息的表
print("\n🔍 Looking for decision-related tables...")
for table_name in tables:
    table_name = table_name[0]
    if 'decision' in table_name.lower() or 'trade' in table_name.lower():
        print(f"\nTable: {table_name}")
        cursor.execute(f'PRAGMA table_info({table_name})')
        columns = cursor.fetchall()
        print("Columns:")
        for col in columns:
            print(f"  - {col[1]} ({col[2]})")

conn.close()