#!/usr/bin/env python3
"""
检查 trader_orders 表中的订单动作类型统计
"""

import sqlite3
import os

def main():
    db_path = os.path.join('data', 'data.db')
    
    if not os.path.exists(db_path):
        print(f'数据库文件不存在: {db_path}')
        return
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    print('订单动作类型统计:')
    cursor.execute('SELECT order_action, COUNT(*) FROM trader_orders GROUP BY order_action ORDER BY COUNT(*) DESC')
    stats = cursor.fetchall()
    for action, count in stats:
        print(f'{action}: {count}')
    
    print('\n开仓/平仓订单记录 (最近20条):')
    cursor.execute("""
        SELECT id, trader_id, symbol, order_action, status, 
               filled_quantity, avg_fill_price, created_at 
        FROM trader_orders 
        WHERE order_action LIKE '%open%' OR order_action LIKE '%close%' 
        ORDER BY id DESC 
        LIMIT 20
    """)
    open_close_orders = cursor.fetchall()
    for order in open_close_orders:
        print(order)
    
    print(f'\n总共订单数: {cursor.execute("SELECT COUNT(*) FROM trader_orders").fetchone()[0]}')
    print(f'开仓订单数: {cursor.execute("SELECT COUNT(*) FROM trader_orders WHERE order_action LIKE \'%open%\'").fetchone()[0]}')
    print(f'平仓订单数: {cursor.execute("SELECT COUNT(*) FROM trader_orders WHERE order_action LIKE \'%close%\'").fetchone()[0]}')
    
    conn.close()

if __name__ == "__main__":
    main()