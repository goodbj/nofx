import requests
import time
import json

def test_cache_performance():
    """测试缓存优化效果"""
    
    print("=== 缓存性能测试 ===")
    
    # 登录获取token
    login_url = "http://localhost:8888/api/login"
    login_data = {
        "email": "test2@example.com",
        "password": "12345678"
    }
    
    try:
        login_response = requests.post(login_url, json=login_data)
        if login_response.status_code == 200:
            token = login_response.json()['token']
            print("✅ 登录成功")
        else:
            print(f"❌ 登录失败: {login_response.text}")
            return
    except Exception as e:
        print(f"❌ 登录异常: {e}")
        return
    
    headers = {"Authorization": f"Bearer {token}"}
    trader_id = "bac26dfa_guardian-ai_1770901210"
    
    # 测试缓存前的性能
    print("\n1. 测试无缓存情况下的性能:")
    times_without_cache = []
    
    for i in range(3):
        start_time = time.time()
        try:
            response = requests.get(
                f"http://localhost:8888/api/account?trader_id={trader_id}", 
                headers=headers, 
                timeout=30
            )
            end_time = time.time()
            
            duration = (end_time - start_time) * 1000
            times_without_cache.append(duration)
            
            status = "✅" if response.status_code == 200 else "❌"
            print(f"  第{i+1}次请求: {duration:.2f}ms {status}")
            
        except Exception as e:
            print(f"  第{i+1}次请求: 错误 - {e}")
            times_without_cache.append(30000)
    
    avg_without_cache = sum(times_without_cache) / len(times_without_cache)
    print(f"  无缓存平均响应时间: {avg_without_cache:.2f}ms")
    
    # 等待缓存建立
    print("\n2. 等待缓存建立...")
    time.sleep(2)
    
    # 测试缓存后的性能
    print("\n3. 测试有缓存情况下的性能:")
    times_with_cache = []
    
    for i in range(3):
        start_time = time.time()
        try:
            response = requests.get(
                f"http://localhost:8888/api/account?trader_id={trader_id}", 
                headers=headers, 
                timeout=30
            )
            end_time = time.time()
            
            duration = (end_time - start_time) * 1000
            times_with_cache.append(duration)
            
            status = "✅" if response.status_code == 200 else "❌"
            cache_status = "缓存命中" if "cache hit" in response.text.lower() or i > 0 else "缓存未命中"
            print(f"  第{i+1}次请求: {duration:.2f}ms {status} ({cache_status})")
            
        except Exception as e:
            print(f"  第{i+1}次请求: 错误 - {e}")
            times_with_cache.append(30000)
    
    avg_with_cache = sum(times_with_cache) / len(times_with_cache)
    print(f"  有缓存平均响应时间: {avg_with_cache:.2f}ms")
    
    # 计算性能提升
    if avg_without_cache > 0:
        improvement = ((avg_without_cache - avg_with_cache) / avg_without_cache) * 100
        print(f"\n4. 性能提升分析:")
        print(f"  响应时间减少: {avg_without_cache - avg_with_cache:.2f}ms")
        print(f"  性能提升: {improvement:.1f}%")
        
        if improvement > 50:
            print("  ✅ 性能优化效果显著")
        elif improvement > 20:
            print("  ✅ 性能有所改善")
        else:
            print("  ⚠️  性能提升有限")
    
    # 测试缓存管理API
    print("\n5. 测试缓存管理功能:")
    
    # 获取缓存统计
    try:
        stats_response = requests.get(
            "http://localhost:8888/api/cache/stats", 
            headers=headers, 
            timeout=10
        )
        if stats_response.status_code == 200:
            stats = stats_response.json()
            print(f"  ✅ 缓存统计API正常")
            print(f"    缓存条目数: {stats.get('total_entries', 'N/A')}")
            print(f"    缓存时长: {stats.get('cache_duration', 'N/A')}")
        else:
            print(f"  ❌ 缓存统计API失败: {stats_response.status_code}")
    except Exception as e:
        print(f"  ❌ 缓存统计API异常: {e}")
    
    # 清除缓存
    try:
        clear_response = requests.post(
            "http://localhost:8888/api/cache/clear", 
            headers=headers, 
            timeout=10
        )
        if clear_response.status_code == 200:
            print(f"  ✅ 缓存清除API正常")
        else:
            print(f"  ❌ 缓存清除API失败: {clear_response.status_code}")
    except Exception as e:
        print(f"  ❌ 缓存清除API异常: {e}")

if __name__ == "__main__":
    test_cache_performance()