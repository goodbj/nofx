import requests
import json
import time
import pyotp

def test_with_real_otp():
    print("=== 使用真实OTP代码测试 ===")
    
    # 1. 注册用户
    print("\n1. 注册用户...")
    register_data = {
        "email": "test2@example.com",
        "password": "password123"
    }
    
    try:
        response = requests.post('http://localhost:8888/api/register', 
                               json=register_data,
                               headers={'Content-Type': 'application/json'})
        print(f"注册状态码: {response.status_code}")
        
        if response.status_code == 200:
            register_result = response.json()
            print(f"注册成功: {register_result.get('message')}")
            user_id = register_result.get('user_id')
            otp_secret = register_result.get('otp_secret')
            print(f"User ID: {user_id}")
            print(f"OTP Secret: {otp_secret}")
            
            # 2. 生成真实的OTP代码
            print(f"\n2. 生成OTP代码...")
            totp = pyotp.TOTP(otp_secret)
            current_otp = totp.now()
            print(f"当前OTP代码: {current_otp}")
            
            # 3. 验证OTP
            print("\n3. 验证OTP...")
            otp_data = {
                "user_id": user_id,
                "otp_code": current_otp
            }
            
            response = requests.post('http://localhost:8888/api/complete-registration',
                                   json=otp_data,
                                   headers={'Content-Type': 'application/json'})
            print(f"OTP验证状态码: {response.status_code}")
            
            if response.status_code == 200:
                login_result = response.json()
                token = login_result.get('token')
                print(f"登录成功，获取令牌: {token[:20]}...")
                
                # 4. 获取交易员列表
                print("\n4. 获取交易员列表...")
                headers = {'Authorization': f'Bearer {token}'}
                response = requests.get('http://localhost:8888/api/my-traders', headers=headers)
                print(f"交易员列表状态码: {response.status_code}")
                
                if response.status_code == 200:
                    traders = response.json()
                    print(f"交易员数量: {len(traders)}")
                    for trader in traders:
                        print(f"  - {trader.get('trader_name')} (ID: {trader.get('trader_id')}, 运行中: {trader.get('is_running')})")
                    
                    # 5. 获取第一个交易员的决策记录
                    if traders:
                        trader_id = traders[0].get('trader_id')
                        print(f"\n5. 获取交易员 {trader_id} 的决策记录...")
                        response = requests.get(f'http://localhost:8888/api/decisions/latest?trader_id={trader_id}&limit=5', 
                                              headers=headers)
                        print(f"决策记录状态码: {response.status_code}")
                        
                        if response.status_code == 200:
                            decisions = response.json()
                            print(f"决策记录数量: {len(decisions)}")
                            for i, decision in enumerate(decisions):
                                print(f"  {i+1}. Cycle {decision.get('cycle_number')}, 时间: {decision.get('timestamp')[:19]}, 成功: {decision.get('success')}")
                        else:
                            print(f"获取决策记录失败: {response.text}")
                else:
                    print(f"获取交易员列表失败: {response.text}")
            else:
                print(f"OTP验证失败: {response.text}")
        else:
            print(f"注册失败: {response.text}")
            
    except Exception as e:
        print(f"流程测试失败: {e}")

if __name__ == "__main__":
    test_with_real_otp()