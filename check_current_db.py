import sqlite3

def check_current_db_structure():
    """检查当前数据库表结构"""
    db_path = 'data/data.db'
    
    print(f"=== 检查数据库: {db_path} ===")
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # 检查所有表
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = cursor.fetchall()
        print(f"表总数: {len(tables)}")
        print("表列表:")
        for table in tables:
            print(f"  - {table[0]}")
        
        # 检查关键表结构
        key_tables = ['trader_positions', 'trader_orders', 'trader_fills']
        for table_name in key_tables:
            print(f"\n--- {table_name} 表结构 ---")
            try:
                cursor.execute(f'PRAGMA table_info({table_name})')
                columns = cursor.fetchall()
                if columns:
                    for col in columns:
                        print(f"  {col[1]} ({col[2]}) {'NOT NULL' if col[3] else 'NULL'}")
                else:
                    print(f"  表不存在")
            except Exception as e:
                print(f"  检查失败: {e}")
        
        # 检查记录数
        print(f"\n--- 记录统计 ---")
        for table_name in key_tables:
            try:
                cursor.execute(f'SELECT COUNT(*) FROM {table_name}')
                count = cursor.fetchone()[0]
                print(f"  {table_name}: {count} 条记录")
            except Exception as e:
                print(f"  {table_name}: 无法查询 ({e})")
        
        conn.close()
        
    except Exception as e:
        print(f"数据库连接失败: {e}")

if __name__ == "__main__":
    check_current_db_structure()