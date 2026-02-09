import sqlite3
import os
from datetime import datetime

def analyze_database_purpose():
    """分析各个数据库文件的作用"""
    
    db_files = ['data/data.db', 'data/nofx.db', 'data/traders.db']
    
    print("=== 数据库文件作用分析 ===\n")
    
    for db_path in db_files:
        if not os.path.exists(db_path):
            print(f"❌ {db_path}: 文件不存在")
            continue
            
        file_size = os.path.getsize(db_path)
        file_time = datetime.fromtimestamp(os.path.getmtime(db_path))
        
        print(f"📊 数据库文件: {db_path}")
        print(f"   文件大小: {file_size:,} 字节 ({file_size/1024/1024:.1f} MB)")
        print(f"   最后修改: {file_time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        try:
            conn = sqlite3.connect(db_path)
            cursor = conn.cursor()
            
            # 获取所有表
            cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
            tables = [row[0] for row in cursor.fetchall()]
            
            print(f"   表数量: {len(tables)}")
            
            # 分析关键表的数据量
            key_tables = ['traders', 'trader_positions', 'trader_orders', 'trader_fills']
            table_stats = {}
            
            for table in key_tables:
                if table in tables:
                    try:
                        cursor.execute(f"SELECT COUNT(*) FROM {table}")
                        count = cursor.fetchone()[0]
                        table_stats[table] = count
                    except:
                        table_stats[table] = "查询失败"
                else:
                    table_stats[table] = "表不存在"
            
            # 判断数据库主要用途
            purpose = "未知用途"
            if 'traders' in tables and 'trader_positions' in tables:
                if table_stats.get('trader_positions', 0) > 0:
                    purpose = "【主数据库】当前生产环境，包含活跃交易数据"
                elif table_stats.get('trader_orders', 0) > 0:
                    purpose = "【历史数据库】包含历史订单记录但无仓位记录"
                else:
                    purpose = "【空数据库】表结构完整但无实际数据"
            
            print(f"   主要用途: {purpose}")
            
            # 显示关键表统计
            print("   关键表记录数:")
            for table, count in table_stats.items():
                status = "✅" if isinstance(count, int) and count > 0 else "❌"
                print(f"     {status} {table}: {count}")
            
            conn.close()
            
        except Exception as e:
            print(f"   ❌ 无法连接数据库: {e}")
        
        print()

if __name__ == "__main__":
    analyze_database_purpose()