import sqlite3

def check_detailed_error_info():
    """检查详细的错误信息"""
    
    print("=== 详细错误信息分析 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        # 获取完整的错误信息
        cursor.execute("""
            SELECT 
                cycle_number,
                timestamp,
                error_message,
                LENGTH(error_message) as error_length
            FROM decision_records 
            WHERE error_message IS NOT NULL 
            AND error_message != ''
            ORDER BY timestamp DESC
            LIMIT 1
        """)
        
        error_record = cursor.fetchone()
        if error_record:
            print(f"\n🔍 完整错误信息 (周期 {error_record[0]}, 时间 {error_record[1][:19]}):")
            print("-" * 80)
            print(error_record[2])
            print("-" * 80)
            print(f"错误信息长度: {error_record[3]} 字符")
            
            # 分析错误类型
            error_msg = error_record[2]
            if "guardian browser automation failed" in error_msg:
                print("\n🎯 错误类型分析:")
                print("  ❌ Guardian浏览器自动化失败")
                if "failed to retrieve" in error_msg:
                    print("  - 浏览器无法获取页面内容")
                if "timeout" in error_msg:
                    print("  - 操作超时")
                if "navigation" in error_msg:
                    print("  - 页面导航失败")
                    
            if "AI API call failed" in error_msg:
                print("  ❌ AI API调用失败")
                
        else:
            print("✅ 无错误记录")
            
        # 检查交易员配置
        print(f"\n=== 交易员配置检查 ===")
        cursor.execute("""
            SELECT 
                name,
                ai_model_id,
                exchange_id,
                scan_interval_minutes,
                custom_prompt,
                override_base_prompt
            FROM traders 
            WHERE is_running = 1
        """)
        
        trader_config = cursor.fetchone()
        if trader_config:
            print(f"交易员名称: {trader_config[0]}")
            print(f"AI模型: {trader_config[1]}")
            print(f"交易所: {trader_config[2]}")
            print(f"扫描间隔: {trader_config[3]} 分钟")
            print(f"自定义提示词: {'是' if trader_config[4] else '否'}")
            print(f"覆盖基础提示词: {'是' if trader_config[5] else '否'}")
            
            # 检查AI模型配置
            print(f"\n=== AI模型配置检查 ===")
            cursor.execute("""
                SELECT name, provider, model, is_active
                FROM ai_models 
                WHERE id = ?
            """, (trader_config[1],))
            
            ai_model = cursor.fetchone()
            if ai_model:
                print(f"模型名称: {ai_model[0]}")
                print(f"提供商: {ai_model[1]}")
                print(f"模型ID: {ai_model[2]}")
                print(f"是否激活: {'是' if ai_model[3] else '否'}")
            else:
                print("❌ 未找到AI模型配置")
                
        conn.close()
        
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_detailed_error_info()