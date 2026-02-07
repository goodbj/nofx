# 检查 PositionHistory 组件中的 Sync 按钮实现

import re

def check_sync_button_implementation():
    file_path = 'web/src/components/PositionHistory.tsx'
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        print("=== Sync 按钮实现检查 ===\n")
        
        # 检查关键函数和状态
        checks = {
            'handleSyncHistory 函数': 'handleSyncHistory' in content,
            'isSyncing 状态': 'isSyncing' in content,
            'syncPositionHistory API 调用': 'syncPositionHistory' in content,
            'Sync 按钮文本': ('Sync' in content or 'sync' in content),
            '按钮渲染': '<button' in content and ('Sync' in content or 'sync' in content)
        }
        
        print("实现状态检查:")
        for check_name, exists in checks.items():
            status = "✅ 存在" if exists else "❌ 缺失"
            print(f"  {status} {check_name}")
        
        # 查找按钮相关代码
        print("\n按钮相关代码片段:")
        lines = content.split('\n')
        for i, line in enumerate(lines, 1):
            if 'button' in line.lower() and ('sync' in line.lower() or 'Sync' in line):
                print(f"  行 {i}: {line.strip()}")
        
        # 查找 handleSyncHistory 函数
        print("\nhandleSyncHistory 函数:")
        func_match = re.search(r'function\s+handleSyncHistory\s*\(.*?\)\s*{.*?}', content, re.DOTALL)
        if func_match:
            func_code = func_match.group(0)
            print("  函数已实现:")
            print("  " + func_code.split('\n')[0])  # 只显示函数签名
        else:
            print("  ❌ 函数未找到")
            
        # 检查 API 调用
        print("\nAPI 调用检查:")
        if 'api.syncPositionHistory' in content:
            print("  ✅ 找到 syncPositionHistory API 调用")
        else:
            print("  ❌ 未找到 syncPositionHistory API 调用")
            
        # 检查翻译键
        print("\n国际化支持:")
        if 'positionHistory.sync' in content:
            print("  ✅ 找到同步功能的翻译键")
        else:
            print("  ❌ 未找到同步功能的翻译键")
            
    except FileNotFoundError:
        print(f"❌ 文件未找到: {file_path}")
    except Exception as e:
        print(f"❌ 检查过程中出错: {e}")

if __name__ == "__main__":
    check_sync_button_implementation()