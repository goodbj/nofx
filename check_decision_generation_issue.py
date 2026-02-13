import sqlite3
from datetime import datetime, timedelta

def check_decision_records_generation():
    """检查决策记录生成情况"""
    
    print("=== 决策记录生成情况检查 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        # 1. 检查运行中交易员的配置
        print("\n1. 运行中交易员配置检查:")
        cursor.execute("""
            SELECT id, name, scan_interval_minutes, is_running, updated_at
            FROM traders 
            WHERE is_running = 1
        """)
        
        running_traders = cursor.fetchall()
        if running_traders:
            for trader in running_traders:
                print(f"  交易员: {trader[1]} (ID: {trader[0]})")
                print(f"    扫描间隔: {trader[2]} 分钟")
                print(f"    运行状态: {'运行中' if trader[3] else '停止'}")
                print(f"    更新时间: {trader[4]}")
        else:
            print("  ❌ 无运行中的交易员")
            return
        
        # 2. 检查该交易员的决策记录
        trader_id = running_traders[0][0]  # 取第一个运行中的交易员
        print(f"\n2. 交易员 {trader_id} 的决策记录检查:")
        
        cursor.execute("""
            SELECT 
                cycle_number,
                timestamp,
                success,
                error_message,
                LENGTH(decision_json) as decision_length
            FROM decision_records 
            WHERE trader_id = ?
            ORDER BY timestamp DESC
            LIMIT 10
        """, (trader_id,))
        
        records = cursor.fetchall()
        print(f"  最近10条记录:")
        for record in records:
            status = "✅" if record[2] else "❌"
            error_info = f" (错误: {record[3][:50]}...)" if record[3] else ""
            print(f"    周期 {record[0]} | {record[1][:19]} | {status}{error_info}")
        
        # 3. 统计决策记录
        print(f"\n3. 决策记录统计:")
        cursor.execute("""
            SELECT 
                COUNT(*) as total_records,
                COUNT(CASE WHEN success = 1 THEN 1 END) as success_count,
                COUNT(CASE WHEN success = 0 THEN 1 END) as failed_count,
                MAX(cycle_number) as max_cycle,
                MIN(timestamp) as first_record,
                MAX(timestamp) as last_record
            FROM decision_records 
            WHERE trader_id = ?
        """, (trader_id,))
        
        stats = cursor.fetchone()
        print(f"  总记录数: {stats[0]}")
        print(f"  成功记录: {stats[1]}")
        print(f"  失败记录: {stats[2]}")
        print(f"  最大周期号: {stats[3]}")
        print(f"  首条记录: {stats[4][:19] if stats[4] else '无'}")
        print(f"  最新记录: {stats[5][:19] if stats[5] else '无'}")
        
        # 4. 检查最近的策略执行情况
        print(f"\n4. 最近1小时策略执行检查:")
        one_hour_ago = datetime.now() - timedelta(hours=1)
        cursor.execute("""
            SELECT 
                cycle_number,
                timestamp,
                success,
                error_message
            FROM decision_records 
            WHERE trader_id = ?
            AND timestamp > ?
            ORDER BY timestamp DESC
        """, (trader_id, one_hour_ago))
        
        recent_records = cursor.fetchall()
        print(f"  最近1小时新增记录数: {len(recent_records)}")
        for record in recent_records:
            status = "✅" if record[2] else "❌"
            error_info = f" (错误: {record[3][:30]}...)" if record[3] else ""
            print(f"    周期 {record[0]} | {record[1][:19]} | {status}{error_info}")
        
        # 5. 检查可能的问题原因
        print(f"\n5. 问题诊断:")
        if stats[0] == 1:
            print("  ⚠️  只有1条记录，可能原因:")
            print("    - 策略执行失败")
            print("    - 决策记录保存失败") 
            print("    - 交易员配置问题")
            print("    - 系统异常中断")
            
            # 检查最近的错误记录
            cursor.execute("""
                SELECT error_message, timestamp
                FROM decision_records 
                WHERE trader_id = ?
                AND success = 0
                AND error_message IS NOT NULL
                AND error_message != ''
                ORDER BY timestamp DESC
                LIMIT 5
            """, (trader_id,))
            
            error_records = cursor.fetchall()
            if error_records:
                print(f"  🔍 最近错误信息:")
                for error in error_records:
                    print(f"    {error[1][:19]}: {error[0][:100]}...")
            else:
                print("  ✅ 无明显错误记录")
        
        conn.close()
        
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_decision_records_generation()