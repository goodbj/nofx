import sqlite3
import time
import datetime

def monitor_decision_generation():
    print("=== 决策生成实时监控 ===")
    print("按 Ctrl+C 停止监控")
    print("-" * 50)
    
    last_count = 0
    last_time = None
    
    while True:
        try:
            # 连接数据库
            conn = sqlite3.connect('data/data.db')
            cursor = conn.cursor()
            
            # 获取当前记录数和最新记录
            cursor.execute("""
                SELECT 
                    count(*) as total_count,
                    max(timestamp) as last_record,
                    max(cycle_number) as current_cycle
                FROM decision_records 
                WHERE trader_id = 'bac26dfa_guardian-ai_1770901210'
            """)
            
            result = cursor.fetchone()
            current_count = result[0]
            last_record = result[1][:19] if result[1] else "无"
            current_cycle = result[2] if result[2] else 0
            
            # 计算时间间隔
            current_time = datetime.datetime.now()
            if last_time and current_count > last_count:
                interval = (current_time - last_time).total_seconds() / 60
                print(f"[{current_time.strftime('%H:%M:%S')}] 新增记录! 间隔: {interval:.1f}分钟")
            
            # 显示当前状态
            print(f"[{current_time.strftime('%H:%M:%S')}] 总记录: {current_count}, 最新周期: {current_cycle}, 最新时间: {last_record}")
            
            # 更新状态
            if current_count > last_count:
                last_time = current_time
                last_count = current_count
            
            conn.close()
            
            # 等待30秒
            time.sleep(30)
            
        except KeyboardInterrupt:
            print("\n监控已停止")
            break
        except Exception as e:
            print(f"错误: {e}")
            time.sleep(10)

if __name__ == "__main__":
    monitor_decision_generation()