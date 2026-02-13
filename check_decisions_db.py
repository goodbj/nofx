import sqlite3

# 连接数据库
conn = sqlite3.connect('nofx.db')
cursor = conn.cursor()

# 检查decisions表结构
print("🔍 Decisions table structure:")
cursor.execute('PRAGMA table_info(decisions)')
for row in cursor.fetchall():
    print(row)

# 检查是否有实际的决策数据
print("\n📊 Sample decision records:")
cursor.execute('SELECT id, decision_json FROM decisions LIMIT 3')
records = cursor.fetchall()
for record in records:
    print(f"ID: {record[0]}")
    print(f"Decision JSON: {record[1][:200]}...")  # 只显示前200字符
    print("-" * 50)

conn.close()