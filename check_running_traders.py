import sqlite3

# 连接数据库
conn = sqlite3.connect('data/data.db')
cursor = conn.cursor()

print("=== 运行中交易员详细配置 ===")

# 获取运行中交易员的详细配置
cursor.execute("""
    SELECT 
        id,
        name,
        scan_interval_minutes,
        strategy_id,
        ai_model_id,
        exchange_id,
        is_running,
        updated_at
    FROM traders 
    WHERE is_running = 1
""")

running_traders = cursor.fetchall()

for trader in running_traders:
    print(f"\n交易员: {trader[1]} (ID: {trader[0]})")
    print(f"  扫描间隔: {trader[2]} 分钟")
    print(f"  策略ID: {trader[3]}")
    print(f"  AI模型: {trader[4]}")
    print(f"  交易所: {trader[5]}")
    print(f"  运行状态: {'运行中' if trader[6] else '停止'}")
    print(f"  更新时间: {trader[7]}")
    
    # 检查该交易员最近的决策记录
    cursor.execute("""
        SELECT 
            count(*) as recent_count,
            max(timestamp) as last_record,
            max(cycle_number) as current_cycle
        FROM decision_records 
        WHERE trader_id = ?
        AND timestamp > datetime('now', '-2 hours')
    """, (trader[0],))
    
    recent_stats = cursor.fetchone()
    print(f"  最近2小时决策记录数: {recent_stats[0]}")
    print(f"  最新记录时间: {recent_stats[1][:19] if recent_stats[1] else '无'}")
    print(f"  当前周期号: {recent_stats[2] if recent_stats[2] else '无'}")

conn.close()