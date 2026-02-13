import sqlite3

def check_guardian_ai_config():
    """检查Guardian AI配置"""
    
    print("=== Guardian AI配置检查 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        # 检查guardian-ai模型配置
        cursor.execute("""
            SELECT 
                id, 
                name, 
                provider, 
                enabled,
                api_key,
                custom_api_url,
                custom_model_name,
                use_browser_automation
            FROM ai_models 
            WHERE id = 'guardian-ai'
        """)
        
        model_config = cursor.fetchone()
        if model_config:
            print(f"模型ID: {model_config[0]}")
            print(f"模型名称: {model_config[1]}")
            print(f"提供商: {model_config[2]}")
            print(f"是否启用: {'是' if model_config[3] else '否'}")
            print(f"API密钥: {'已设置' if model_config[4] else '未设置'}")
            print(f"自定义API URL: {model_config[5] if model_config[5] else '无'}")
            print(f"自定义模型名: {model_config[6] if model_config[6] else '无'}")
            print(f"使用浏览器自动化: {'是' if model_config[7] else '否'}")
        else:
            print("❌ 未找到guardian-ai模型配置")
            
        # 检查Guardian相关配置
        print(f"\n=== 浏览器自动化配置 ===")
        cursor.execute("""
            SELECT key, value 
            FROM system_configs 
            WHERE key LIKE '%guardian%' OR key LIKE '%browser%'
        """)
        
        guardian_configs = cursor.fetchall()
        if guardian_configs:
            for config in guardian_configs:
                print(f"  {config[0]}: {config[1]}")
        else:
            print("  无相关配置")
            
        conn.close()
        
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_guardian_ai_config()