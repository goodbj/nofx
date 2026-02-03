import sqlite3
import os

def check_trader_data():
    # 连接数据库
    db_path = './data/data.db'
    if not os.path.exists(db_path):
        print(f"数据库文件不存在: {db_path}")
        return
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    print("=== 检查交易员表结构 ===")
    
    # 查看表结构
    cursor.execute("PRAGMA table_info(traders)")
    columns = cursor.fetchall()
    
    print("traders表字段:")
    for col in columns:
        cid, name, typ, notnull, dflt_value, pk = col
        print(f"  {cid}. {name} ({typ}) NOT NULL:{notnull} PK:{pk} DEFAULT:{dflt_value}")
    
    print("\n=== 检查交易员数据 ===")
    
    # 查询所有交易员的数据
    cursor.execute("""
        SELECT id, name, data_access_method, exchange_id, ai_model_id, is_running
        FROM traders 
        ORDER BY created_at DESC
    """)
    
    traders = cursor.fetchall()
    
    if not traders:
        print("没有找到任何交易员数据")
        return
    
    print(f"\n总共找到 {len(traders)} 个交易员记录:")
    
    for i, trader in enumerate(traders, 1):
        trader_id, name, data_access_method, exchange_id, ai_model_id, is_running = trader
        print(f"\n{i}. 交易员信息:")
        print(f"   ID: {trader_id}")
        print(f"   名称: {name}")
        print(f"   数据访问方式: '{data_access_method}'")
        print(f"   交易所ID: {exchange_id}")
        print(f"   AI模型ID: {ai_model_id}")
        print(f"   是否运行中: {is_running}")
        
        # 检查data_access_method的值
        if data_access_method == 'native':
            print("   ⚠️  当前设置为不使用代理")
        elif data_access_method == 'proxy':
            print("   ✅ 当前设置为使用代理")
        else:
            print(f"   ❓ 未知的数据访问方式: '{data_access_method}'")
    
    conn.close()

if __name__ == "__main__":
    check_trader_data()