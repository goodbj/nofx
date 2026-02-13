import sqlite3
import os

def check_decision_data_source():
    """检查dashboard页面周期内容的数据来源"""
    
    # 检查数据库文件
    db_path = 'data/data.db'
    if not os.path.exists(db_path):
        print(f"❌ 数据库文件不存在: {db_path}")
        return
    
    print("=== Dashboard周期内容数据源检查 ===")
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # 1. 检查decision_records表是否存在
        print("\n1. 检查decision_records表结构:")
        cursor.execute("PRAGMA table_info(decision_records)")
        columns = cursor.fetchall()
        
        if columns:
            print("✅ decision_records表存在，列结构:")
            for col in columns:
                print(f"   - {col[1]} ({col[2]})")
        else:
            print("❌ decision_records表不存在")
            return
        
        # 2. 检查表中是否有数据
        print("\n2. 检查数据记录:")
        cursor.execute("SELECT COUNT(*) FROM decision_records")
        total_count = cursor.fetchone()[0]
        print(f"   总记录数: {total_count}")
        
        if total_count > 0:
            # 3. 查看最近的记录
            print("\n3. 最近5条记录预览:")
            cursor.execute("""
                SELECT 
                    id,
                    trader_id,
                    cycle_number,
                    timestamp,
                    success,
                    error_message
                FROM decision_records 
                ORDER BY timestamp DESC 
                LIMIT 5
            """)
            
            records = cursor.fetchall()
            for record in records:
                print(f"   ID: {record[0]}, Trader: {record[1][:20]}..., Cycle: {record[2]}, Time: {record[3][:19]}, Success: {record[4]}")
                if record[5]:
                    print(f"     Error: {record[5][:50]}...")
            
            # 4. 检查特定交易员的数据
            print("\n4. 检查运行中交易员的数据:")
            cursor.execute("""
                SELECT id, name, is_running 
                FROM traders 
                WHERE is_running = 1 
                LIMIT 1
            """)
            running_trader = cursor.fetchone()
            
            if running_trader:
                trader_id = running_trader[0]
                trader_name = running_trader[1]
                print(f"   运行中交易员: {trader_name} (ID: {trader_id})")
                
                cursor.execute("""
                    SELECT COUNT(*) 
                    FROM decision_records 
                    WHERE trader_id = ?
                """, (trader_id,))
                trader_records = cursor.fetchone()[0]
                print(f"   该交易员决策记录数: {trader_records}")
                
                if trader_records > 0:
                    cursor.execute("""
                        SELECT cycle_number, timestamp, success
                        FROM decision_records 
                        WHERE trader_id = ?
                        ORDER BY timestamp DESC
                        LIMIT 3
                    """, (trader_id,))
                    
                    trader_data = cursor.fetchall()
                    print("   最近3条记录:")
                    for data in trader_data:
                        print(f"     Cycle {data[0]} | {data[1][:19]} | {'成功' if data[2] else '失败'}")
            else:
                print("   无运行中的交易员")
        
        # 5. 检查数据完整性
        print("\n5. 数据完整性检查:")
        cursor.execute("SELECT COUNT(DISTINCT trader_id) FROM decision_records")
        unique_traders = cursor.fetchone()[0]
        print(f"   涉及交易员数量: {unique_traders}")
        
        cursor.execute("SELECT MAX(cycle_number) FROM decision_records")
        max_cycle = cursor.fetchone()[0]
        print(f"   最大周期号: {max_cycle}")
        
        cursor.execute("""
            SELECT 
                COUNT(*) as total,
                COUNT(CASE WHEN success = 1 THEN 1 END) as success,
                COUNT(CASE WHEN success = 0 THEN 1 END) as failed
            FROM decision_records
        """)
        stats = cursor.fetchone()
        print(f"   成功记录: {stats[1]}, 失败记录: {stats[2]}, 总计: {stats[0]}")
        
        conn.close()
        
        # 6. 总结
        print("\n=== 数据源总结 ===")
        print("✅ Dashboard页面的'周期'内容存储在:")
        print("   - 数据库: data/data.db")
        print("   - 表名: decision_records")
        print("   - 关键字段: cycle_number, timestamp, trader_id, success")
        print("   - 数据通过API: /api/decisions/latest 获取")
        print("   - 前端组件: DecisionCard.tsx 显示")
        
    except Exception as e:
        print(f"❌ 数据库查询出错: {e}")

if __name__ == "__main__":
    check_decision_data_source()