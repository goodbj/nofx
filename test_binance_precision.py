#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
币安精度问题修复验证脚本
测试AKEUSDT的精度处理是否正确
"""

import requests
import json
import sys

def get_symbol_info(symbol):
    """获取交易对的精度信息"""
    try:
        url = f"https://fapi.binance.com/fapi/v1/exchangeInfo?symbol={symbol}"
        response = requests.get(url, timeout=10)
        data = response.json()
        
        if 'symbols' in data and len(data['symbols']) > 0:
            symbol_info = data['symbols'][0]
            print(f"📊 {symbol} 交易对信息:")
            print(f"   状态: {symbol_info.get('status', 'N/A')}")
            print(f"   基础资产: {symbol_info.get('baseAsset', 'N/A')}")
            print(f"   报价资产: {symbol_info.get('quoteAsset', 'N/A')}")
            
            # 获取LOT_SIZE过滤器
            lot_size_filter = None
            price_filter = None
            
            for filter_info in symbol_info.get('filters', []):
                if filter_info.get('filterType') == 'LOT_SIZE':
                    lot_size_filter = filter_info
                elif filter_info.get('filterType') == 'PRICE_FILTER':
                    price_filter = filter_info
            
            print(f"\n🔢 数量精度信息 (LOT_SIZE):")
            if lot_size_filter:
                print(f"   minQty: {lot_size_filter.get('minQty', 'N/A')}")
                print(f"   maxQty: {lot_size_filter.get('maxQty', 'N/A')}")
                print(f"   stepSize: {lot_size_filter.get('stepSize', 'N/A')}")
                
                # 计算精度
                step_size = lot_size_filter.get('stepSize', '1')
                if '.' in step_size:
                    precision = len(step_size.split('.')[1])
                    print(f"   计算精度: {precision} 位小数")
                else:
                    print(f"   计算精度: 0 位小数")
            else:
                print("   未找到LOT_SIZE过滤器")
            
            print(f"\n💰 价格精度信息 (PRICE_FILTER):")
            if price_filter:
                print(f"   minPrice: {price_filter.get('minPrice', 'N/A')}")
                print(f"   maxPrice: {price_filter.get('maxPrice', 'N/A')}")
                print(f"   tickSize: {price_filter.get('tickSize', 'N/A')}")
                
                # 计算价格精度
                tick_size = price_filter.get('tickSize', '1')
                if '.' in tick_size:
                    price_precision = len(tick_size.split('.')[1])
                    print(f"   价格精度: {price_precision} 位小数")
                else:
                    print(f"   价格精度: 0 位小数")
            else:
                print("   未找到PRICE_FILTER过滤器")
            
            return lot_size_filter, price_filter
        else:
            print(f"❌ 未找到 {symbol} 的交易对信息")
            return None, None
            
    except Exception as e:
        print(f"❌ 获取 {symbol} 信息失败: {e}")
        return None, None

def test_quantity_formatting():
    """测试数量格式化逻辑"""
    print("\n🧪 数量格式化测试:")
    
    test_cases = [
        ("AKEUSDT", 1.23456789),  # 高精度数量
        ("AKEUSDT", 0.001),       # 最小数量
        ("AKEUSDT", 0.0015),      # 需要对齐到stepSize
        ("BTCUSDT", 0.001234),    # BTC测试
        ("ETHUSDT", 0.012345),    # ETH测试
    ]
    
    for symbol, quantity in test_cases:
        print(f"\n   测试 {symbol} 数量 {quantity}:")
        
        # 获取精度信息
        lot_filter, _ = get_symbol_info(symbol)
        if lot_filter:
            step_size = float(lot_filter.get('stepSize', '1'))
            min_qty = float(lot_filter.get('minQty', '0'))
            
            # 模拟步长对齐
            aligned_qty = (quantity // step_size) * step_size
            print(f"      stepSize: {step_size}")
            print(f"      原始数量: {quantity}")
            print(f"      对齐后数量: {aligned_qty}")
            print(f"      是否满足最小数量: {aligned_qty >= min_qty}")

def main():
    """主函数"""
    print("🔍 币安精度问题诊断工具")
    print("=" * 50)
    
    # 测试目标交易对
    target_symbol = "AKEUSDT"
    
    # 获取并显示精度信息
    lot_filter, price_filter = get_symbol_info(target_symbol)
    
    # 测试格式化逻辑
    test_quantity_formatting()
    
    print("\n" + "=" * 50)
    print("✅ 诊断完成")
    
    if lot_filter:
        step_size = lot_filter.get('stepSize')
        print(f"📋 AKEUSDT关键信息:")
        print(f"   stepSize: {step_size}")
        print(f"   精度要求: {len(step_size.split('.')[1]) if '.' in step_size else 0} 位小数")
        print(f"   修复建议: 确保所有下单数量是 {step_size} 的整数倍")

if __name__ == "__main__":
    main()