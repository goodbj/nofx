import sqlite3

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== traders表结构 ===")
cursor.execute("PRAGMA table_info(traders)")
columns = cursor.fetchall()
for col in columns:
    print(f"列名: {col[1]}, 类型: {col[2]}, 非空: {col[3]}, 默认值: {col[4]}, 主键: {col[5]}")

print("\n=== 交易员当前状态 ===")
cursor.execute("SELECT * FROM traders WHERE id = 'bac26dfa_guardian-ai_1770901210'")
trader_data = cursor.fetchone()
if trader_data:
    print("交易员数据:")
    for i, col in enumerate(columns):
        print(f"  {col[1]}: {trader_data[i]}")

conn.close()