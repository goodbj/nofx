import sqlite3

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

# 获取交易员及其用户信息
print("=== 交易员与用户关联 ===")
cursor.execute("""
    SELECT t.id, t.name, t.user_id, t.is_running, u.email 
    FROM traders t 
    LEFT JOIN users u ON t.user_id = u.id
""")
traders = cursor.fetchall()

for trader in traders:
    print(f"交易员ID: {trader[0]}")
    print(f"  名称: {trader[1]}")
    print(f"  用户ID: {trader[2]}")
    print(f"  用户邮箱: {trader[4] or 'N/A'}")
    print(f"  运行状态: {'运行中' if trader[3] else '停止'}")
    print()

conn.close()