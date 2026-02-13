import sqlite3
import requests
import json

def check_trader_status_consistency():
    """检查交易员状态一致性"""
    
    print("=== 交易员状态一致性检查 ===")
    
    # 1. 检查数据库状态
    print("\n1. 数据库中的交易员状态:")
    try:
        conn = sqlite3.connect('data/data.db')
        cursor = conn.cursor()
        
        cursor.execute('SELECT id, name, is_running, updated_at FROM traders')
        db_traders = cursor.fetchall()
        
        db_status = {}
        for trader in db_traders:
            status_text = "运行中" if trader[2] else "已停止"
            print(f"  {trader[1]} (ID: {trader[0]}) - 状态: {status_text}, 更新时间: {trader[3]}")
            db_status[trader[0]] = trader[2]  # 存储数据库状态
        
        conn.close()
    except Exception as e:
        print(f"  ❌ 检查数据库时出错: {e}")
        return
    
    # 2. 检查内存状态 (通过API)
    print("\n2. 内存中的交易员状态 (通过API):")
    
    # 先登录获取token
    try:
        login_response = requests.post("http://localhost:8888/api/login", json={
            "email": "test2@example.com",
            "password": "12345678"
        })
        
        if login_response.status_code != 200:
            print(f"  ❌ 登录失败: {login_response.text}")
            return
            
        token = login_response.json()['token']
        headers = {"Authorization": f"Bearer {token}"}
        
        # 获取我的交易员列表
        traders_response = requests.get("http://localhost:8888/api/my-traders", headers=headers)
        if traders_response.status_code != 200:
            print(f"  ❌ 获取交易员列表失败: {traders_response.text}")
            return
            
        traders_list = traders_response.json()
        
        for trader in traders_list:
            trader_id = trader['trader_id']
            trader_name = trader['name']
            is_running_api = trader.get('is_running', False)
            is_running_text = "运行中" if is_running_api else "已停止"
            
            # 获取详细状态
            status_response = requests.get(f"http://localhost:8888/api/status/{trader_id}", headers=headers)
            if status_response.status_code == 200:
                status_details = status_response.json()
                is_running_detail = status_details.get('is_running', False)
                is_executing = status_details.get('is_executing', False)
                runtime_minutes = status_details.get('runtime_minutes', 0)
                
                print(f"  {trader_name} (ID: {trader_id}):")
                print(f"    API列表状态: {is_running_text}")
                print(f"    详细状态: {'运行中' if is_running_detail else '已停止'}")
                print(f"    正在执行: {'是' if is_executing else '否'}")
                print(f"    运行时长: {runtime_minutes} 分钟")
                
                # 检查一致性
                db_running = bool(db_status.get(trader_id, False))
                api_running = is_running_detail
                
                if db_running != api_running:
                    print(f"    ⚠️  状态不一致! 数据库: {'运行' if db_running else '停止'}, 内存/API: {'运行' if api_running else '停止'}")
                else:
                    print(f"    ✅ 状态一致")
            else:
                print(f"  {trader_name} (ID: {trader_id}) - 状态获取失败")
                
    except Exception as e:
        print(f"  ❌ 检查API状态时出错: {e}")

if __name__ == "__main__":
    check_trader_status_consistency()