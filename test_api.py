import requests
import json

# 1. 首先尝试管理员登录
print("=== 尝试管理员登录 ===")
try:
    response = requests.post('http://localhost:8888/api/admin-login', 
                           json={'password': 'admin'}, 
                           headers={'Content-Type': 'application/json'})
    print(f"状态码: {response.status_code}")
    print(f"响应: {response.text}")
    
    if response.status_code == 200:
        data = response.json()
        token = data.get('token')
        print(f"获取到令牌: {token[:20]}...")
        
        # 2. 使用令牌获取交易员列表
        print("\n=== 获取交易员列表 ===")
        headers = {'Authorization': f'Bearer {token}'}
        response = requests.get('http://localhost:8888/api/my-traders', headers=headers)
        print(f"状态码: {response.status_code}")
        print(f"响应: {response.text}")
        
        if response.status_code == 200:
            traders = response.json()
            print(f"交易员数量: {len(traders)}")
            for trader in traders:
                print(f"  - {trader.get('trader_name')} (ID: {trader.get('trader_id')})")
                
            # 3. 获取第一个交易员的决策记录
            if traders:
                trader_id = traders[0].get('trader_id')
                print(f"\n=== 获取交易员 {trader_id} 的决策记录 ===")
                response = requests.get(f'http://localhost:8888/api/decisions/latest?trader_id={trader_id}&limit=5', 
                                      headers=headers)
                print(f"状态码: {response.status_code}")
                print(f"响应: {response.text}")
                
    else:
        print("管理员登录失败")
        
except Exception as e:
    print(f"请求失败: {e}")