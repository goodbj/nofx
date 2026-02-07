import sqlite3
import os

def check_nofx_db():
    try:
        # 检查 nofx.db
        print("=== 检查 nofx.db ===")
        conn = sqlite3.connect('data/nofx.db')
        cursor = conn.cursor()
        
        # 检查所有表
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = cursor.fetchall()
        print("nofx.db 中的表:", [t[0] for t in tables])
        
        # 检查关键表
        key_tables = ['trader_positions', 'trader_orders', 'trader_fills']
        for table in key_tables:
            if (table,) in tables:
                cursor.execute(f"SELECT COUNT(*) FROM {table}")
                count = cursor.fetchone()[0]
                print(f"✅ {table}: {count} 条记录")
            else:
                print(f"❌ {table}: 表不存在")
        
        conn.close()
        
        # 检查 traders.db
        print("\n=== 检查 traders.db ===")
        conn2 = sqlite3.connect('data/traders.db')
        cursor2 = conn2.cursor()
        
        cursor2.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables2 = cursor2.fetchall()
        print("traders.db 中的表:", [t[0] for t in tables2])
        
        # 检查 traders 表
        if ('traders',) in tables2:
            cursor2.execute("SELECT COUNT(*) FROM traders")
            trader_count = cursor2.fetchone()[0]
            print(f"✅ traders 表: {trader_count} 条记录")
        else:
            print("❌ traders 表不存在")
            
        conn2.close()
        
    except Exception as e:
        print(f"数据库检查错误: {e}")

if __name__ == "__main__":
    check_nofx_db()