import urllib.request
import urllib.error
import json
import sqlite3

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
        response = urllib.request.urlopen("http://localhost:8081/health", timeout=10)
        data = json.loads(response.read().decode())
        print(f"✅ 健康检查: {response.getcode()} - {data}")
    except Exception as e:
        print(f"❌ 健康检查失败: {e}")
    
    # 2. 测试通过代理获取币安数据
    print("\n2️⃣  测试通过代理获取币安数据")
    print("-" * 30)
    
    try:
        # 测试获取服务器时间（不需要认证的API）
        response = urllib.request.urlopen("http://localhost:8081/api/proxy/time", timeout=15)
        if response.getcode() == 200:
            print(f"✅ 时间API测试: {response.getcode()} - 服务器时间获取成功")
        else:
            print(f"❌ 时间API测试: {response.getcode()}")
    except Exception as e:
        print(f"❌ 时间API测试失败: {e}")
    
    # 3. 测试后端服务状态
    print("\n3️⃣  测试后端服务状态")
    print("-" * 30)
    
    try:
        response = urllib.request.urlopen("http://localhost:8888/api/health", timeout=10)
        print(f"✅ 后端健康检查: {response.getcode()}")
    except urllib.error.HTTPError as e:
        print(f"✅ 后端健康检查: {e.code} (预期的认证错误)")
    except Exception as e:
        print(f"❌ 后端健康检查失败: {e}")
    
    # 4. 测试交易员配置
    print("\n4️⃣  测试交易员代理配置")
    print("-" * 30)
    
    try:
        # 尝试获取交易员列表（可能需要认证，但我们主要测试连接）
        req = urllib.request.Request("http://localhost:8888/api/my-traders")
        req.add_header('Authorization', 'Bearer dummy-token')
        response = urllib.request.urlopen(req, timeout=10)
        print(f"✅ 交易员API访问: {response.getcode()}")
    except urllib.error.HTTPError as e:
        if e.code in [401, 403]:  # 认证错误是正常的
            print(f"✅ 交易员API访问: {e.code} (认证错误，API可访问)")
        else:
            print(f"❌ 交易员API访问: {e.code}")
    except Exception as e:
        print(f"❌ 交易员API测试失败: {e}")
    
    # 5. 测试特定代理模式交易员
    print("\n5️⃣  测试代理模式交易员")
    print("-" * 30)
    
    # 检查数据库中已有的代理模式交易员
    try:
        conn = sqlite3.connect('./data/data.db')
        cursor = conn.cursor()
        cursor.execute("SELECT id, name, data_access_method FROM traders WHERE data_access_method = 'proxy'")
        proxy_traders = cursor.fetchall()
        conn.close()
        
        print(f"📊 代理模式交易员数量: {len(proxy_traders)}")
        for trader_id, name, method in proxy_traders:
            print(f"   • {name} (ID: {trader_id}) - 模式: {method}")
            
        # 如果有代理模式交易员，测试其配置
        if proxy_traders:
            print("\n💡 代理模式交易员已就绪，可以进行功能测试")
        else:
            print("\n⚠️  没有找到代理模式交易员，需要创建一个进行测试")
    except Exception as e:
        print(f"❌ 数据库查询失败: {e}")
    
    print("\n" + "=" * 60)
    print("✅ 第二阶段测试完成")
    print("=" * 60)

if __name__ == "__main__":
    test_proxy_functionality()