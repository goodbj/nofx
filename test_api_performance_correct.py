import requests
import time
import json

def test_api_performance_with_correct_user():
    """使用正确用户测试API性能"""
    
    print("=== API性能测试 ===")
    
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
    
    # 测试不同的API端点
    test_endpoints = [
        ("/api/traders", "交易员列表"),
        ("/api/account?trader_id=bac26dfa_guardian-ai_1770901210", "账户信息"),
        ("/api/positions?trader_id=bac26dfa_guardian-ai_1770901210", "持仓信息"),
        ("/api/decisions/latest?trader_id=bac26dfa_guardian-ai_1770901210&limit=5", "最新决策"),
        ("/api/statistics?trader_id=bac26dfa_guardian-ai_1770901210", "统计数据")
    ]
    
    results = []
    
    for endpoint, description in test_endpoints:
        print(f"\n测试 {description} ({endpoint})")
        
        # 多次测试取平均值
        times = []
        for i in range(3):
            start_time = time.time()
            try:
                response = requests.get(f"http://localhost:8888{endpoint}", headers=headers, timeout=30)
                end_time = time.time()
                
                duration = (end_time - start_time) * 1000  # 转换为毫秒
                times.append(duration)
                
                status = "✅" if response.status_code == 200 else "❌"
                print(f"  第{i+1}次: {duration:.2f}ms {status}")
                
                if response.status_code != 200:
                    print(f"    错误: {response.text}")
                    
            except Exception as e:
                print(f"  第{i+1}次: 错误 - {e}")
                times.append(30000)  # 30秒超时
        
        if times:
            avg_time = sum(times) / len(times)
            max_time = max(times)
            min_time = min(times)
            results.append({
                "endpoint": endpoint,
                "description": description,
                "avg_time": avg_time,
                "max_time": max_time,
                "min_time": min_time
            })
            print(f"  平均响应时间: {avg_time:.2f}ms")
    
    # 输出性能报告
    print("\n=== 性能测试报告 ===")
    print(f"{'接口描述':<15} {'平均时间(ms)':<15} {'最慢(ms)':<15} {'最快(ms)':<15}")
    print("-" * 60)
    
    for result in results:
        print(f"{result['description']:<15} {result['avg_time']:<15.2f} {result['max_time']:<15.2f} {result['min_time']:<15.2f}")
    
    # 识别性能瓶颈
    slow_endpoints = [r for r in results if r['avg_time'] > 1000]  # 超过1秒的认为是慢的
    if slow_endpoints:
        print(f"\n⚠️  发现性能瓶颈:")
        for endpoint in slow_endpoints:
            print(f"  - {endpoint['description']}: {endpoint['avg_time']:.2f}ms")
    else:
        print(f"\n✅ 所有接口响应时间都在合理范围内")

if __name__ == "__main__":
    test_api_performance_with_correct_user()