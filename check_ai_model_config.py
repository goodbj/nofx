import sqlite3

def check_ai_model_config():
    """检查AI模型配置"""
    
    print("=== AI模型配置检查 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        # 检查guardian-ai模型配置
        cursor.execute("""
            SELECT id, name, provider, model, is_active, api_key, base_url
            FROM ai_models 
            WHERE id = 'guardian-ai'
        """)
        
        model_config = cursor.fetchone()
        if model_config:
            print(f"模型ID: {model_config[0]}")
            print(f"模型名称: {model_config[1]}")
            print(f"提供商: {model_config[2]}")
            print(f"模型: {model_config[3]}")
            print(f"是否激活: {'是' if model_config[4] else '否'}")
            print(f"API密钥: {model_config[5][:10] if model_config[5] else '无'}...")
            print(f"基础URL: {model_config[6] if model_config[6] else '无'}")
        else:
            print("❌ 未找到guardian-ai模型配置")
            
        # 检查所有AI模型
        print(f"\n=== 所有AI模型 ===")
        cursor.execute("""
            SELECT id, name, provider, is_active
            FROM ai_models
        """)
        
        all_models = cursor.fetchall()
        for model in all_models:
            status = "✅" if model[3] else "❌"
            print(f"  {status} {model[0]} ({model[1]}) - {model[2]}")
            
        conn.close()
        
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_ai_model_config()