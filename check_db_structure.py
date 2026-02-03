import sqlite3
import sys

def check_database():
    try:
        conn = sqlite3.connect('traders.db')
        cursor = conn.cursor()
        
        # 检查所有表
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table';")
        tables = cursor.fetchall()
        print("数据库中的表:", tables)
        
        if ('traders',) in tables:
            # 检查traders表的结构
            cursor.execute("PRAGMA table_info(traders);")
            columns = cursor.fetchall()
            print("\ntraders表结构:")
            for col in columns:
                print(f"  {col[1]} ({col[2]}) - {'NOT NULL' if col[3] else 'NULL'}, Default: {col[4]}")
            
            # 检查是否有data_access_method列
            column_names = [col[1] for col in columns]
            if 'data_access_method' in column_names:
                print("\n✅ data_access_method 字段存在")
                
                # 查看一些示例数据
                cursor.execute("SELECT id, name, data_access_method FROM traders LIMIT 5;")
                rows = cursor.fetchall()
                print(f"\n示例数据 (最多5条):")
                for row in rows:
                    print(f"  {row[0][:8]}... - {row[1]} - {row[2]}")
            else:
                print("\n❌ data_access_method 字段不存在")
        else:
            print("\n❌ traders 表不存在")
        
        conn.close()
    except Exception as e:
        print(f"数据库错误: {e}")
        # 检查是否文件不存在
        import os
        if not os.path.exists('traders.db'):
            print("traders.db 文件不存在")

if __name__ == "__main__":
    check_database()