import sqlite3
import os

def check_all_databases():
    """检查所有可能的数据库文件"""
    data_dir = 'data'
    if not os.path.exists(data_dir):
        print("数据目录不存在")
        return
    
    # 列出所有数据库文件
    db_files = [f for f in os.listdir(data_dir) if f.endswith('.db')]
    print("=== 检查所有数据库文件 ===")
    
    for db_file in db_files:
        db_path = os.path.join(data_dir, db_file)
        print(f"\n--- 检查数据库: {db_file} ---")
        
        try:
            conn = sqlite3.connect(db_path)
            cursor = conn.cursor()
            
            # 获取所有表
            cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
            tables = cursor.fetchall()
            print(f"表数量: {len(tables)}")
            
            # 检查关键交易表
            key_tables = ['trader_positions', 'trader_orders', 'trader_fills', 'traders']
            for table in key_tables:
                if (table,) in tables:
                    cursor.execute(f"SELECT COUNT(*) FROM {table}")
                    count = cursor.fetchone()[0]
                    print(f"✅ {table}: {count} 条记录")
                else:
                    print(f"❌ {table}: 表不存在")
            
            conn.close()
        except Exception as e:
            print(f"❌ 检查 {db_file} 时出错: {e}")

if __name__ == "__main__":
    check_all_databases()