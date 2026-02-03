import sqlite3
import uuid
from datetime import datetime

def update_trader_proxy_setting():
    # 连接数据库
    db_path = './data/data.db'
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    print("=== 更新交易员代理设置测试 ===")
    
    # 查看当前所有交易员
    cursor.execute("""
        SELECT id, name, data_access_method 
        FROM traders 
        ORDER BY created_at DESC
    """)
    
    traders = cursor.fetchall()
    
    if not traders:
        print("没有找到任何交易员数据")
        return
    
    print(f"\n找到 {len(traders)} 个交易员:")
    for i, trader in enumerate(traders, 1):
        trader_id, name, data_access_method = trader
        print(f"{i}. {name} (ID: {trader_id}) - 当前设置: {data_access_method}")
    
    # 选择第一个交易员进行测试
    if traders:
        test_trader_id = traders[0][0]
        test_trader_name = traders[0][1]
        
        print(f"\n=== 准备更新交易员 ===")
        print(f"交易员: {test_trader_name}")
        print(f"ID: {test_trader_id}")
        print(f"当前设置: native")
        print(f"目标设置: proxy")
        
        # 更新为proxy模式
        cursor.execute("""
            UPDATE traders 
            SET data_access_method = 'proxy', updated_at = ?
            WHERE id = ?
        """, (datetime.now().strftime('%Y-%m-%d %H:%M:%S'), test_trader_id))
        
        conn.commit()
        
        print(f"\n✅ 成功更新交易员 {test_trader_name} 的代理设置为 'proxy'")
        
        # 验证更新结果
        cursor.execute("""
            SELECT id, name, data_access_method 
            FROM traders 
            WHERE id = ?
        """, (test_trader_id,))
        
        updated_trader = cursor.fetchone()
        if updated_trader:
            trader_id, name, data_access_method = updated_trader
            print(f"\n=== 验证结果 ===")
            print(f"交易员: {name}")
            print(f"ID: {trader_id}")
            print(f"更新后的设置: {data_access_method}")
            
            if data_access_method == 'proxy':
                print("✅ 代理设置更新成功!")
            else:
                print("❌ 代理设置更新失败!")
    
    conn.close()

if __name__ == "__main__":
    update_trader_proxy_setting()