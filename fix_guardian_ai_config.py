import sqlite3

def fix_guardian_ai_config():
    """修复Guardian AI配置问题"""
    
    print("=== Guardian AI配置修复 ===")
    
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        print("当前配置:")
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
            
            # 检查是否需要修复
            if current_config[4] and current_config[5] == 1:
                print("\n⚠️  检测到配置冲突:")
                print("  - 已设置自定义API URL")
                print("  - 但浏览器自动化仍为开启状态")
                print("  - 这会导致浏览器访问API URL失败")
                
                # 修复配置：关闭浏览器自动化
                print("\n🔧 正在修复配置...")
                cursor.execute("""
                    UPDATE ai_models 
                    SET use_browser_automation = 0,
                        updated_at = datetime('now')
                    WHERE id = 'guardian-ai'
                """)
                
                if cursor.rowcount > 0:
                    conn.commit()
                    print("✅ 已关闭浏览器自动化，现在将直接使用API调用")
                    
                    # 验证修复结果
                    cursor.execute("""
                        SELECT use_browser_automation 
                        FROM ai_models 
                        WHERE id = 'guardian-ai'
                    """)
                    result = cursor.fetchone()
                    print(f"验证结果: 浏览器自动化 = {'开启' if result[0] else '关闭'}")
                else:
                    print("❌ 更新失败")
            else:
                print("\n✅ 配置看起来正常，无需修复")
                
        else:
            print("❌ 未找到guardian-ai模型配置")
            
        conn.close()
        
    except Exception as e:
        print(f"❌ 修复过程中出错: {e}")

def show_fix_options():
    """显示修复选项"""
    print("\n=== 修复选项说明 ===")
    print("根据您的配置，有两种解决方案:")
    print()
    print("1. 🎯 API直接调用模式（推荐）")
    print("   - 保持当前的 https://chat.deepseek.com 配置")
    print("   - 关闭浏览器自动化")
    print("   - 优势：更快、更稳定")
    print()
    print("2. 🌐 浏览器访问模式")
    print("   - 清空自定义API URL")
    print("   - 开启浏览器自动化")
    print("   - 优势：可以使用网页界面")
    print()
    print("当前脚本已自动选择选项1进行修复")

if __name__ == "__main__":
    show_fix_options()
    fix_guardian_ai_config()