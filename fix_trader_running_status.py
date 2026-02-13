import sqlite3

def fix_trader_running_status():
    """修复交易员运行状态问题"""
    
    print("=== 交易员运行状态修复 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        print("\n1. 当前数据库状态:")
        cursor.execute('SELECT id, name, is_running FROM traders')
        traders = cursor.fetchall()
        
        for trader in traders:
            status_text = "运行中" if trader[2] else "已停止"
            print(f"  {trader[1]} (ID: {trader[0]}) - 状态: {status_text}")
        
        print("\n2. 更新所有交易员状态为停止...")
        
        # 更新所有交易员为停止状态
        cursor.execute('UPDATE traders SET is_running = 0')
        rows_affected = cursor.rowcount
        print(f"  ✅ 已更新 {rows_affected} 个交易员的状态为停止")
        
        print("\n3. 更新后的状态:")
        cursor.execute('SELECT id, name, is_running FROM traders')
        updated_traders = cursor.fetchall()
        
        for trader in updated_traders:
            status_text = "运行中" if trader[2] else "已停止"
            print(f"  {trader[1]} (ID: {trader[0]}) - 状态: {status_text}")
        
        conn.commit()
        conn.close()
        
        print(f"\n✅ 修复完成！所有交易员现在都处于停止状态。")
        print("💡 提示：需要重启后端服务才能使更改生效。")
        
    except Exception as e:
        print(f"❌ 修复过程中出错: {e}")

if __name__ == "__main__":
    fix_trader_running_status()