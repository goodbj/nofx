#!/usr/bin/env python3
"""
验证 trader_positions 表是否已正确填充
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
    
    # 检查 trader_positions 表的记录数
    cursor.execute('SELECT COUNT(*) FROM trader_positions')
    count = cursor.fetchone()[0]
    print(f'trader_positions 表记录数: {count}')
    
    # 检查开仓和已平仓的记录数
    cursor.execute('SELECT status, COUNT(*) FROM trader_positions GROUP BY status')
    statuses = cursor.fetchall()
    for status, cnt in statuses:
        print(f'状态 {status}: {cnt} 条记录')
    
    # 检查特定交易员的仓位
    cursor.execute('SELECT COUNT(*) FROM trader_positions WHERE trader_id = "d84ca79f_guardian-ai_1770428140"')
    guardian_positions = cursor.fetchone()[0]
    print(f'Guardian AI 交易员仓位数: {guardian_positions}')
    
    # 检查最近的几个仓位记录
    print('\n最近的5个仓位记录:')
    cursor.execute('''
        SELECT id, symbol, side, status, quantity, entry_price, exit_price, 
               entry_time, exit_time
        FROM trader_positions 
        ORDER BY id DESC 
        LIMIT 5
    ''')
    records = cursor.fetchall()
    for record in records:
        print(record)
    
    # 检查是否有开放的仓位
    print('\n开放仓位 (OPEN):')
    cursor.execute('''
        SELECT id, symbol, side, quantity, entry_price, 
               datetime(entry_time/1000, "unixepoch", "localtime") as entry_time
        FROM trader_positions 
        WHERE status = "OPEN"
        ORDER BY entry_time DESC
    ''')
    open_positions = cursor.fetchall()
    for pos in open_positions:
        print(pos)
    
    conn.close()

if __name__ == "__main__":
    main()