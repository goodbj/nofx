#!/usr/bin/env python3
"""
从已有的订单记录中同步仓位信息到 trader_positions 表
"""

import sqlite3
import os
import sys

def main():
    db_path = os.path.join('data', 'data.db')
    
    if not os.path.exists(db_path):
        print(f'数据库文件不存在: {db_path}')
        return
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # 首先检查当前的仓位表状态
    cursor.execute('SELECT COUNT(*) FROM trader_positions')
    current_positions = cursor.fetchone()[0]
    print(f'当前 trader_positions 表记录数: {current_positions}')
    
    # 检查是否有开仓订单
    cursor.execute("""
        SELECT id, trader_id, exchange_id, exchange_type, symbol, 
               order_action, filled_quantity, avg_fill_price, 
               filled_at as entry_time, exchange_order_id
        FROM trader_orders 
        WHERE order_action LIKE '%open%' 
        ORDER BY filled_at ASC
    """)
    open_orders = cursor.fetchall()
    
    print(f'找到 {len(open_orders)} 个开仓订单')
    
    # 检查对应的平仓订单
    cursor.execute("""
        SELECT id, trader_id, exchange_id, exchange_type, symbol, 
               order_action, filled_quantity, avg_fill_price, 
               filled_at as exit_time, exchange_order_id
        FROM trader_orders 
        WHERE order_action LIKE '%close%' 
        ORDER BY filled_at ASC
    """)
    close_orders = cursor.fetchall()
    
    print(f'找到 {len(close_orders)} 个平仓订单')
    
    # 检查是否存在 trader_positions 表的结构
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='trader_positions'")
    if not cursor.fetchone():
        print('错误: trader_positions 表不存在')
        conn.close()
        return
    
    # 尝试插入开仓记录
    inserted_count = 0
    for order in open_orders:
        order_id, trader_id, exchange_id, exchange_type, symbol, order_action, \
        quantity, price, entry_time, exchange_order_id = order
        
        # 确定仓位方向
        side = "LONG" if "long" in order_action else "SHORT"
        
        try:
            # 检查是否已存在相同的仓位记录
            cursor.execute("""
                SELECT id FROM trader_positions 
                WHERE trader_id = ? AND symbol = ? AND side = ? AND status = 'OPEN'
            """, (trader_id, symbol, side))
            
            existing = cursor.fetchone()
            if existing:
                print(f'  跳过已存在的开仓记录: {symbol} {side} for {trader_id}')
                continue
            
            # 插入新的开仓记录
            cursor.execute("""
                INSERT INTO trader_positions (
                    trader_id, exchange_id, exchange_type, exchange_position_id,
                    symbol, side, quantity, entry_price, entry_order_id, entry_time,
                    leverage, status, source, created_at, updated_at, entry_quantity
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                trader_id, exchange_id or '', exchange_type or '', exchange_order_id,
                symbol, side, quantity, price, exchange_order_id, entry_time,
                1, 'OPEN', 'order_sync', entry_time, entry_time, quantity
            ))
            inserted_count += 1
            
        except sqlite3.Error as e:
            print(f'  插入仓位记录失败: {e}')
    
    # 处理平仓记录
    closed_count = 0
    for order in close_orders:
        order_id, trader_id, exchange_id, exchange_type, symbol, order_action, \
        quantity, price, exit_time, exchange_order_id = order
        
        # 确定仓位方向
        side = "LONG" if "long" in order_action else "SHORT"
        
        try:
            # 查找对应的开仓记录并更新为已平仓
            cursor.execute("""
                UPDATE trader_positions 
                SET exit_price = ?, exit_order_id = ?, exit_time = ?, 
                    status = 'CLOSED', close_reason = ?, 
                    updated_at = ?
                WHERE trader_id = ? AND symbol = ? AND side = ? AND status = 'OPEN'
                ORDER BY entry_time DESC LIMIT 1
            """, (
                price, exchange_order_id, exit_time,
                f'manual_sync_{order_action}', exit_time,
                trader_id, symbol, side
            ))
            
            if cursor.rowcount > 0:
                closed_count += 1
                
        except sqlite3.Error as e:
            print(f'  更新平仓记录失败: {e}')
    
    conn.commit()
    conn.close()
    
    print(f'\n同步完成!')
    print(f'新增开仓记录: {inserted_count}')
    print(f'更新平仓记录: {closed_count}')
    
    # 重新检查结果
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute('SELECT COUNT(*) FROM trader_positions')
    new_count = cursor.fetchone()[0]
    print(f'同步后 trader_positions 表记录数: {new_count}')
    conn.close()

if __name__ == "__main__":
    main()