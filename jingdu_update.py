#!/usr/bin/env python3
# jingdu_update.py - 精度文件更新工具

import json
import requests
import time
from datetime import datetime

def update_jingdu_json():
    """更新jingdu.json文件"""
    print("🔍 开始更新jingdu.json文件...")
    
    # 获取最新exchangeInfo
    try:
        response = requests.get(
            "https://testnet.binancefuture.com/fapi/v1/exchangeInfo",
            timeout=30
        )
        response.raise_for_status()
        exchange_info = response.json()
        
        print(f"✅ 成功获取API数据，服务器时间: {exchange_info['serverTime']}")
        print(f"✅ 对应时间: {datetime.fromtimestamp(exchange_info['serverTime']/1000)}")
        
        # 准备写入的数据格式
        data_to_write = {
            "serverTime": exchange_info["serverTime"],
            "symbols": []
        }
        
        # 转换symbol格式
        for symbol in exchange_info["symbols"]:
            # 简化数据结构，只保留必要的字段
            simplified_symbol = {
                "symbol": symbol["symbol"],
                "pricePrecision": symbol["pricePrecision"],
                "quantityPrecision": symbol["quantityPrecision"],
                "filters": []
            }
            
            # 转换过滤器
            for filter_item in symbol["filters"]:
                filter_type = filter_item["filterType"]
                new_filter = {"filterType": filter_type}
                
                if filter_type == "PRICE_FILTER":
                    new_filter["tickSize"] = filter_item.get("tickSize", "")
                elif filter_type == "LOT_SIZE":
                    new_filter["stepSize"] = filter_item.get("stepSize", "")
                    new_filter["minQty"] = filter_item.get("minQty", "")
                    new_filter["maxQty"] = filter_item.get("maxQty", "")
                    
                simplified_symbol["filters"].append(new_filter)
                
            data_to_write["symbols"].append(simplified_symbol)
            
        # 写入文件
        backup_path = f"data/jingdu.json.backup.{int(time.time())}"
        
        # 创建备份
        try:
            with open("data/jingdu.json", "r", encoding="utf-8") as f:
                backup_content = f.read()
            with open(backup_path, "w", encoding="utf-8") as f:
                f.write(backup_content)
            print(f"✅ 创建备份文件: {backup_path}")
        except Exception as e:
            print(f"⚠️ 备份创建失败: {e}")
            
        # 更新主文件
        with open("data/jingdu.json", "w", encoding="utf-8") as f:
            json.dump(data_to_write, f, indent=2, ensure_ascii=False)
            
        print("✅ 文件更新成功!")
        print(f"📊 更新后的服务器时间: {data_to_write['serverTime']}")
        print(f"📊 更新时间: {datetime.fromtimestamp(data_to_write['serverTime']/1000)}")
        print(f"📊 包含交易对数量: {len(data_to_write['symbols'])}")
        
        return True
        
    except Exception as e:
        print(f"❌ 更新失败: {e}")
        return False

if __name__ == "__main__":
    update_jingdu_json()