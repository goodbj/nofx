#!/usr/bin/env python3
"""
统一更新项目中所有端口配置的脚本
"""

import os
import re
import sys
from pathlib import Path

def update_ports(new_ports):
    """
    统一更新项目中所有端口配置
    new_ports: dict, 例如 {"BINANCE_PROXY_PORT": "8081", "BINANCE_PROXY_INTERNAL_PORT": "8082"}
    """
    project_root = Path(__file__).parent.parent
    
    # 定义需要更新的文件和端口正则表达式
    files_to_update = [
        {
            "path": project_root / ".env",
            "patterns": {
                r"(BINANCE_PROXY_URL=http://localhost:)\d+": lambda port: f"BINANCE_PROXY_URL=http://localhost:{new_ports.get('BINANCE_PROXY_PORT', '8081')}",
            }
        },
        {
            "path": project_root / "docker-compose.binance-proxy.yml",
            "patterns": {
                r'(\s*-\s*"\d+:\d+")': lambda port: f'      - "{new_ports.get("BINANCE_PROXY_PORT", "8081")}:{new_ports.get("BINANCE_PROXY_INTERNAL_PORT", "8082")}"',
                r'(PORT=)\d+': lambda port: f'PORT={new_ports.get("BINANCE_PROXY_INTERNAL_PORT", "8082")}',
            }
        },
        {
            "path": project_root / "docker-compose.proxy.yml",
            "patterns": {
                r'(\s*-\s*"\d+:\d+")': lambda port: f'      - "{new_ports.get("BINANCE_PROXY_PORT", "8081")}:{new_ports.get("BINANCE_PROXY_INTERNAL_PORT", "8082")}"',
                r'(PORT=)\d+': lambda port: f'PORT={new_ports.get("BINANCE_PROXY_INTERNAL_PORT", "8082")}',
            }
        },
        {
            "path": project_root / "START_ALL_SERVICES.bat",
            "patterns": {
                r'(docker run -d --name binance-proxy -p \d+):(\d+)': lambda host, container: f'docker run -d --name binance-proxy -p {new_ports.get("BINANCE_PROXY_PORT", "8081")}:{new_ports.get("BINANCE_PROXY_INTERNAL_PORT", "8082")}',
            }
        },
    ]
    
    for file_info in files_to_update:
        filepath = file_info["path"]
        if not filepath.exists():
            print(f"文件不存在: {filepath}")
            continue
            
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original_content = content
        
        # 应用每个模式
        for pattern, replacement_func in file_info["patterns"].items():
            def replace_match(match):
                groups = match.groups()
                if groups:
                    return replacement_func(*groups)
                else:
                    return replacement_func(match.group())
                    
            content = re.sub(pattern, replace_match, content)
        
        # 如果内容发生了变化，写回文件
        if content != original_content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"已更新: {filepath}")
        else:
            print(f"无变化: {filepath}")

def main():
    if len(sys.argv) < 3:
        print("使用方法: python update_ports.py <外部端口> <内部端口>")
        print("例如: python update_ports.py 8081 8082")
        return
    
    external_port = sys.argv[1]
    internal_port = sys.argv[2]
    
    print(f"更新端口配置:")
    print(f"  外部端口: {external_port}")
    print(f"  内部端口: {internal_port}")
    
    new_ports = {
        "BINANCE_PROXY_PORT": external_port,
        "BINANCE_PROXY_INTERNAL_PORT": internal_port
    }
    
    update_ports(new_ports)
    print("\n端口配置更新完成!")

if __name__ == "__main__":
    main()