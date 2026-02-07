# Sync 按钮可见性测试脚本

import requests
import json

def test_sync_button_visibility():
    """测试 Sync 按钮相关功能"""
    
    print("=== Sync 按钮可见性测试 ===\n")
    
    # 1. 检查 API 端点是否存在
    print("1. 检查 syncPositionHistory API 端点:")
    try:
        # 测试一个假的 traderId 来检查 API 是否存在
        response = requests.get('http://localhost:8888/api/traders/fake-id/sync-history', timeout=5)
        if response.status_code == 404:
            print("  ❌ API 端点不存在")
        elif response.status_code == 400:
            print("  ✅ API 端点存在（返回400是正常的参数错误）")
        else:
            print(f"  ⚠️  API 端点返回状态码: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("  ❌ 无法连接到后端服务，请确保服务正在运行")
    except Exception as e:
        print(f"  ❌ 测试出错: {e}")
    
    # 2. 检查实际的交易员
    print("\n2. 检查可用的交易员:")
    try:
        response = requests.get('http://localhost:8888/api/traders', timeout=5)
        if response.status_code == 200:
            traders = response.json()
            print(f"  ✅ 找到 {len(traders)} 个交易员:")
            for trader in traders:
                print(f"    - {trader.get('name', 'Unknown')} (ID: {trader.get('id', 'Unknown')})")
        else:
            print(f"  ❌ 获取交易员列表失败: {response.status_code}")
    except Exception as e:
        print(f"  ❌ 检查交易员时出错: {e}")
    
    # 3. 检查前端文件
    print("\n3. 检查前端组件文件:")
    import os
    component_path = 'web/src/components/PositionHistory.tsx'
    if os.path.exists(component_path):
        print("  ✅ PositionHistory.tsx 文件存在")
        # 检查关键内容
        with open(component_path, 'r', encoding='utf-8') as f:
            content = f.read()
            checks = {
                'handleSyncHistory 函数': 'handleSyncHistory' in content,
                'syncPositionHistory API': 'syncPositionHistory' in content,
                'Sync 按钮渲染': '<button' in content and 'Sync' in content,
                'isSyncing 状态': 'isSyncing' in content
            }
            for check_name, exists in checks.items():
                status = "✅" if exists else "❌"
                print(f"    {status} {check_name}")
    else:
        print("  ❌ PositionHistory.tsx 文件不存在")
    
    # 4. 检查国际化文件
    print("\n4. 检查国际化翻译:")
    i18n_path = 'web/src/i18n/translations.ts'
    if os.path.exists(i18n_path):
        print("  ✅ translations.ts 文件存在")
        with open(i18n_path, 'r', encoding='utf-8') as f:
            content = f.read()
            if 'positionHistory.sync' in content:
                print("  ✅ 找到 positionHistory.sync 翻译键")
            else:
                print("  ❌ 未找到 positionHistory.sync 翻译键")
    else:
        print("  ❌ translations.ts 文件不存在")
    
    print("\n=== 测试完成 ===")
    print("\n如果所有检查都通过，但按钮仍不可见，请检查:")
    print("1. 是否选择了有效的交易员")
    print("2. 浏览器开发者工具中的控制台错误")
    print("3. 网络请求是否正常")
    print("4. CSS 样式是否有冲突")

if __name__ == "__main__":
    test_sync_button_visibility()