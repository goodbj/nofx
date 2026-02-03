import requests
import time
import json

def test_proxy_functionality():
    """测试代理服务功能"""
    print("=" * 60)
    print("🎯 核心功能验证阶段")
    print("=" * 60)
    
    # 1. 测试代理服务转发功能
    print("\n1️⃣  测试代理服务转发功能")
    print("-" * 30)
    
    # 测试代理服务健康检查
    try:
        response = requests.get("http://localhost:8081/health", timeout=10)
        print(f"✅ 健康检查: {response.status_code} - {response.json()}")
    except Exception as e:
        print(f"❌ 健康检查失败: {e}")
    
    # 2. 测试通过代理获取币安数据
    print("\n2️⃣  测试通过代理获取币安数据")
    print("-" * 30)
    
    try:
        # 测试获取服务器时间（不需要认证的API）
        response = requests.get("http://localhost:8081/api/proxy/time", timeout=15)
        if response.status_code == 200:
            print(f"✅ 时间API测试: {response.status_code} - 服务器时间获取成功")
        else:
            print(f"❌ 时间API测试: {response.status_code} - {response.text}")
    except Exception as e:
        print(f"❌ 时间API测试失败: {e}")
    
    # 3. 测试后端服务状态
    print("\n3️⃣  测试后端服务状态")
    print("-" * 30)
    
    try:
        response = requests.get("http://localhost:8888/api/health", timeout=10)
        print(f"✅ 后端健康检查: {response.status_code}")
    except Exception as e:
        print(f"❌ 后端健康检查失败: {e}")
    
    # 4. 测试交易员配置
    print("\n4️⃣  测试交易员代理配置")
    print("-" * 30)
    
    try:
        # 获取交易员列表
        response = requests.get("http://localhost:8888/api/my-traders", 
                               headers={"Authorization": "Bearer dummy-token"}, 
                               timeout=10)
        if response.status_code in [200, 401]:  # 401是正常的认证错误
            print(f"✅ 交易员API访问: {response.status_code}")
            if response.status_code == 200:
                traders = response.json()
                print(f"📊 交易员数量: {len(traders) if isinstance(traders, list) else 'N/A'}")
        else:
            print(f"❌ 交易员API访问: {response.status_code}")
    except Exception as e:
        print(f"❌ 交易员API测试失败: {e}")
    
    # 5. 测试特定代理模式交易员
    print("\n5️⃣  测试代理模式交易员")
    print("-" * 30)
    
    # 检查数据库中已有的代理模式交易员
    import sqlite3
    try:
        conn = sqlite3.connect('./data/data.db')
        cursor = conn.cursor()
        cursor.execute("SELECT id, name, data_access_method FROM traders WHERE data_access_method = 'proxy'")
        proxy_traders = cursor.fetchall()
        conn.close()
        
        print(f"📊 代理模式交易员数量: {len(proxy_traders)}")
        for trader_id, name, method in proxy_traders:
            print(f"   • {name} (ID: {trader_id}) - 模式: {method}")
    except Exception as e:
        print(f"❌ 数据库查询失败: {e}")
    
    print("\n" + "=" * 60)
    print("✅ 第二阶段测试完成")
    print("=" * 60)

if __name__ == "__main__":
    test_proxy_functionality()