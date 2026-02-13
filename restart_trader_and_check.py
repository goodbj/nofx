import requests
import json
import time

def check_and_restart_trader():
    """检查并重新启动交易员"""
    
    print("=== 交易员状态检查和重启 ===")
    
    # 登录获取token
    login_url = "http://localhost:8888/api/login"
    login_data = {
        "email": "test2@example.com",
        "password": "12345678"
    }
    
    try:
        # 登录
        print("1. 登录系统...")
        login_response = requests.post(login_url, json=login_data)
        if login_response.status_code == 200:
            token = login_response.json()['token']
            print("✅ 登录成功")
        else:
            print(f"❌ 登录失败: {login_response.text}")
            return
            
        headers = {"Authorization": f"Bearer {token}"}
        
        # 获取运行中的交易员
        print("2. 获取运行中的交易员...")
        traders_response = requests.get("http://localhost:8888/api/my-traders", headers=headers)
        if traders_response.status_code == 200:
            traders = traders_response.json()
            running_traders = [t for t in traders if t.get('is_running', False)]
            
            if running_traders:
                trader = running_traders[0]
                trader_id = trader['trader_id']
                trader_name = trader['name']
                print(f"✅ 找到运行中的交易员: {trader_name} (ID: {trader_id})")
                
                # 停止交易员
                print("3. 停止交易员...")
                stop_response = requests.post(
                    f"http://localhost:8888/api/traders/{trader_id}/stop",
                    headers=headers
                )
                if stop_response.status_code == 200:
                    print("✅ 交易员已停止")
                    time.sleep(2)
                else:
                    print(f"⚠️ 停止交易员失败: {stop_response.text}")
                
                # 启动交易员
                print("4. 启动交易员...")
                start_response = requests.post(
                    f"http://localhost:8888/api/traders/{trader_id}/start",
                    headers=headers
                )
                if start_response.status_code == 200:
                    print("✅ 交易员已启动")
                else:
                    print(f"❌ 启动交易员失败: {start_response.text}")
                    return
                    
                # 等待一段时间让交易员运行
                print("5. 等待交易员执行...")
                time.sleep(30)
                
                # 检查决策记录
                print("6. 检查决策记录生成情况...")
                decisions_response = requests.get(
                    f"http://localhost:8888/api/decisions/latest?trader_id={trader_id}&limit=5",
                    headers=headers
                )
                
                if decisions_response.status_code == 200:
                    decisions = decisions_response.json()
                    print(f"✅ 获取到 {len(decisions)} 条决策记录:")
                    for i, decision in enumerate(decisions):
                        status = "✅" if decision.get('success', False) else "❌"
                        print(f"  {i+1}. 周期 {decision.get('cycle_number', 'N/A')} | {decision.get('timestamp', 'N/A')[:19]} | {status}")
                else:
                    print(f"❌ 获取决策记录失败: {decisions_response.text}")
                    
            else:
                print("❌ 无运行中的交易员")
                
        else:
            print(f"❌ 获取交易员列表失败: {traders_response.text}")
            
    except Exception as e:
        print(f"❌ 操作过程中出错: {e}")

if __name__ == "__main__":
    check_and_restart_trader()