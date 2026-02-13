import requests
import json

def test_decision_api_with_correct_user():
    """使用正确用户测试决策API接口"""
    
    print("=== 使用正确用户测试决策API接口 ===")
    
    # 登录信息
    login_url = "http://localhost:8888/api/login"
    login_data = {
        "email": "test@example.com",
        "password": "12345678"
    }
    
    try:
        # 登录获取token
        print("1. 尝试登录获取认证令牌...")
        login_response = requests.post(login_url, json=login_data)
        
        if login_response.status_code == 200:
            login_result = login_response.json()
            if login_result.get('success') and login_result.get('token'):
                token = login_result['token']
                print("✅ 登录成功，获得认证令牌")
            else:
                print("❌ 登录失败:", login_result.get('error', 'Unknown error'))
                return
        else:
            print("❌ 登录请求失败，状态码:", login_response.status_code)
            print("响应内容:", login_response.text)
            return
            
    except Exception as e:
        print("❌ 登录请求异常:", str(e))
        return
    
    # 使用token测试API
    api_url = "http://localhost:8888/api/decisions/latest"
    params = {
        "trader_id": "bac26dfa_guardian-ai_1770901210",
        "limit": 3
    }
    
    headers = {
        "Authorization": f"Bearer {token}"
    }
    
    try:
        print("2. 调用决策API接口...")
        response = requests.get(api_url, params=params, headers=headers)
        
        print("响应状态码:", response.status_code)
        
        if response.status_code == 200:
            data = response.json()
            print("✅ API调用成功")
            print("返回记录数:", len(data))
            
            if len(data) > 0:
                print("\n返回的决策记录:")
                for i, record in enumerate(data):
                    print(f"\n记录 {i+1}:")
                    print(f"  周期号: {record.get('cycle_number', 'N/A')}")
                    print(f"  时间戳: {record.get('timestamp', 'N/A')}")
                    print(f"  成功: {record.get('success', 'N/A')}")
                    print(f"  交易员ID: {record.get('trader_id', 'N/A')[:30]}...")
            else:
                print("返回数据为空")
                
        else:
            print("❌ API调用失败")
            print("错误信息:", response.text)
            
    except Exception as e:
        print("❌ API请求异常:", str(e))

if __name__ == "__main__":
    test_decision_api_with_correct_user()