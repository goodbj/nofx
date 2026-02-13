import requests
import json

# 获取系统配置
print("=== 系统配置 ===")
try:
    response = requests.get('http://localhost:8888/api/config')
    print(f"状态码: {response.status_code}")
    if response.status_code == 200:
        config = response.json()
        print(json.dumps(config, indent=2, ensure_ascii=False))
    else:
        print(f"响应: {response.text}")
except Exception as e:
    print(f"请求失败: {e}")