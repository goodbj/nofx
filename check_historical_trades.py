import sqlite3
import os

def check_historical_trades():
    """检查历史交易记录"""
    # 检查包含交易记录的数据库
    db_files = [
        'data/data20260123更新完测试页测试按钮后的备份.db',
        'data/data20260121添加AI模型前的备份.db'
    ]
    
    for db_file in db_files:
        if os.path.exists(db_file):
            print(f"\n=== 检查数据库: {os.path.basename(db_file)} ===")
            try:
                conn = sqlite3.connect(db_file)
                cursor = conn.cursor()
                
                # 获取最近的仓位记录
                cursor.execute('''
                    SELECT symbol, side, quantity, entry_price, entry_time, status 
                    FROM trader_positions 
                    ORDER BY entry_time DESC 
                    LIMIT 5
                ''')
                positions = cursor.fetchall()
                
                print(f"最近5条仓位记录:")
                for pos in positions:
                    print(f"  {pos[0]} {pos[1]} {pos[2]} @ {pos[3]} (时间: {pos[4]}, 状态: {pos[5]})")
                
                # 获取订单统计
                cursor.execute('SELECT COUNT(*) FROM trader_orders')
                order_count = cursor.fetchone()[0]
                print(f"订单总数: {order_count}")
                
                # 获取成交统计
                cursor.execute('SELECT COUNT(*) FROM trader_fills')
                fill_count = cursor.fetchone()[0]
                print(f"成交总数: {fill_count}")
                
                conn.close()
            except Exception as e:
                print(f"检查失败: {e}")

if __name__ == "__main__":
    check_historical_trades()