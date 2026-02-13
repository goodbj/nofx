import sqlite3

def check_users():
    """检查用户信息"""
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        print("=== 用户信息检查 ===")
        cursor.execute("SELECT email, created_at FROM users LIMIT 10")
        users = cursor.fetchall()
        
        if users:
            print("现有用户:")
            for user in users:
                print(f"  邮箱: {user[0]}, 创建时间: {user[1]}")
        else:
            print("无用户数据")
            
        conn.close()
    except Exception as e:
        print(f"查询用户信息出错: {e}")

if __name__ == "__main__":
    check_users()