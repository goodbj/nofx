import sqlite3

def fix_browser_automation_config():
    """修复浏览器自动化配置以支持token节省"""
    
    print("=== 浏览器自动化配置修复 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        print("当前配置状态:")
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
        
        current_config = cursor.fetchone()
        if current_config:
            print(f"  模型ID: {current_config[0]}")
            print(f"  模型名称: {current_config[1]}")
            print(f"  自定义API URL: {current_config[4] if current_config[4] else '无'}")
            print(f"  浏览器自动化: {'开启' if current_config[5] else '关闭'}")
            
            # 分析当前配置问题
            if current_config[5] == 0:  # 浏览器自动化关闭
                print("\n⚠️  发现问题:")
                print("  - 浏览器自动化当前是关闭状态")
                print("  - 但您希望使用浏览器自动化来节省token")
                
                # 修复配置：开启浏览器自动化，清空自定义API URL
                print("\n🔧 正在修复配置...")
                cursor.execute("""
                    UPDATE ai_models 
                    SET use_browser_automation = 1,
                        custom_api_url = '',
                        updated_at = datetime('now')
                    WHERE id = 'guardian-ai'
                """)
                
                if cursor.rowcount > 0:
                    conn.commit()
                    print("✅ 配置已修复:")
                    print("  - 已开启浏览器自动化")
                    print("  - 已清空自定义API URL")
                    print("  - 现在将使用浏览器访问Guardian AI网页")
                    
                    # 验证修复结果
                    cursor.execute("""
                        SELECT use_browser_automation, custom_api_url
                        FROM ai_models 
                        WHERE id = 'guardian-ai'
                    """)
                    result = cursor.fetchone()
                    print(f"验证结果: 浏览器自动化 = {'开启' if result[0] else '关闭'}")
                    print(f"验证结果: 自定义API URL = '{result[1] if result[1] else '无'}'")
                else:
                    print("❌ 更新失败")
            else:
                print("\n✅ 浏览器自动化已开启，检查其他配置...")
                
                # 如果浏览器自动化已开启但仍有问题，检查API URL
                if current_config[4]:  # 有自定义API URL
                    print("⚠️  检测到配置冲突:")
                    print("  - 浏览器自动化开启")
                    print("  - 但设置了自定义API URL")
                    print("  - 这可能导致浏览器访问错误的URL")
                    
                    print("\n🔧 清理冲突配置...")
                    cursor.execute("""
                        UPDATE ai_models 
                        SET custom_api_url = '',
                            updated_at = datetime('now')
                        WHERE id = 'guardian-ai'
                    """)
                    
                    if cursor.rowcount > 0:
                        conn.commit()
                        print("✅ 已清空自定义API URL，确保浏览器访问正确页面")
                    else:
                        print("❌ 清理失败")
                        
        else:
            print("❌ 未找到guardian-ai模型配置")
            
        conn.close()
        
    except Exception as e:
        print(f"❌ 修复过程中出错: {e}")

def show_configuration_summary():
    """显示配置说明"""
    print("\n=== 配置说明 ===")
    print("为实现浏览器自动化节省token的目标:")
    print("1. ✅ 开启浏览器自动化 (use_browser_automation = 1)")
    print("2. ✅ 清空自定义API URL (让浏览器访问默认的Guardian AI页面)")
    print("3. 🎯 浏览器将直接访问Guardian AI网页界面")
    print("4. 💰 避免API token消耗，通过网页交互获取AI响应")

if __name__ == "__main__":
    show_configuration_summary()
    fix_browser_automation_config()