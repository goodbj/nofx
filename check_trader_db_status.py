import sqlite3

def check_trader_status():
    """检查交易员状态"""
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        print("=== 数据库中的交易员状态 ===")
        cursor.execute('SELECT id, name, is_running, updated_at FROM traders')
        traders = cursor.fetchall()
        
        for trader in traders:
            status_text = "运行中" if trader[2] else "已停止"
            print(f"{trader[1]} (ID: {trader[0]}) - 状态: {status_text}, 更新时间: {trader[3]}")
        
        conn.close()
        
    except Exception as e:
        print(f"检查过程中出错: {e}")

if __name__ == "__main__":
    check_trader_status()