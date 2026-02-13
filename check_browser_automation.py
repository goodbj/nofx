import sqlite3

def check_browser_automation_config():
    """检查浏览器自动化配置"""
    
    print("=== 浏览器自动化配置检查 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        # 检查AI模型配置
        print("1. AI模型配置:")
        cursor.execute("""
            SELECT 
                id, 
                name, 
                provider, 
                enabled,
                custom_api_url,
                use_browser_automation
            FROM ai_models 
            WHERE id = 'guardian-ai'
        """)
        
        model_config = cursor.fetchone()
        if model_config:
            print(f"  模型ID: {model_config[0]}")
            print(f"  模型名称: {model_config[1]}")
            print(f"  提供商: {model_config[2]}")
            print(f"  是否启用: {'是' if model_config[3] else '否'}")
            print(f"  自定义API URL: {model_config[4] if model_config[4] else '无'}")
            print(f"  使用浏览器自动化: {'是' if model_config[5] else '否'}")
            
            # 检查配置是否正确
            if model_config[5] == 1:  # 浏览器自动化开启
                if model_config[4]:  # 有自定义API URL
                    print("  ⚠️  配置警告: 同时启用了浏览器自动化和自定义API URL")
                    print("  这可能导致浏览器尝试访问API URL而不是正常的网页")
                else:
                    print("  ✅ 浏览器自动化配置正常")
            else:
                print("  ❌ 浏览器自动化未开启")
        else:
            print("  ❌ 未找到guardian-ai模型配置")
            
        # 检查系统配置中的浏览器相关设置
        print("\n2. 系统浏览器配置:")
        try:
            cursor.execute("""
                SELECT key, value 
                FROM configs 
                WHERE key LIKE '%browser%' OR key LIKE '%guardian%' OR key LIKE '%chrome%'
            """)
            
            browser_configs = cursor.fetchall()
            if browser_configs:
                for config in browser_configs:
                    print(f"  {config[0]}: {config[1]}")
            else:
                print("  无相关浏览器配置")
        except:
            print("  无法查询configs表")
            
        # 检查最近的错误记录详情
        print("\n3. 详细错误信息分析:")
        cursor.execute("""
            SELECT 
                cycle_number,
                timestamp,
                error_message
            FROM decision_records 
            WHERE error_message LIKE '%browser%' 
            OR error_message LIKE '%guardian%'
            OR error_message LIKE '%automation%'
            ORDER BY timestamp DESC
            LIMIT 3
        """)
        
        error_records = cursor.fetchall()
        if error_records:
            for record in error_records:
                print(f"  周期 {record[0]} | {record[1][:19]}")
                print(f"  错误: {record[2]}")
                print()
        else:
            print("  无浏览器相关错误记录")
            
        conn.close()
        
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_browser_automation_config()