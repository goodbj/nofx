#!/usr/bin/env python3
"""
端口配置验证脚本
验证所有 Docker 配置文件中的端口是否正确引用环境变量
"""

import yaml
import json
import re
from pathlib import Path

def validate_docker_compose_ports():
    """验证 Docker Compose 文件中的端口配置"""
    compose_files = [
        'docker-compose.yml',
        'docker-compose.dev.watch.yml', 
        'docker-compose.dev.yml',
        'docker-compose.prod.yml',
        'docker-compose.stable.yml'
    ]
    
    all_valid = True
    
    for filename in compose_files:
        filepath = Path(filename)
        if not filepath.exists():
            print(f"跳过不存在的文件: {filename}")
            continue
            
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
                
            # 检查是否有硬编码的特定端口（8888, 3300, 8080, 3000）
            hardcoded_ports = re.findall(r'\b(8888|3300|8080|3000)\b(?!\s*:)', content)
            
            # 过滤掉在注释中的端口
            lines = content.split('\n')
            valid_hardcoded = []
            
            for port in hardcoded_ports:
                found_in_comment = False
                for line in lines:
                    # 检查端口是否在注释中
                    comment_pos = line.find('#')
                    if comment_pos != -1:
                        comment_part = line[comment_pos:]
                        port_pos = line.find(port)
                        if port_pos != -1 and port_pos >= comment_pos:
                            # 端口在注释中，有效
                            found_in_comment = True
                            break
                    else:
                        # 没有注释，检查是否在环境变量引用中
                        if f"${{{port}" in line or f":-{port}" in line:
                            # 这是在环境变量引用中，有效
                            found_in_comment = True
                            break
                
                if not found_in_comment:
                    valid_hardcoded.append(port)
            
            if valid_hardcoded:
                print(f"❌ {filename} 中发现硬编码端口: {valid_hardcoded}")
                all_valid = False
            else:
                print(f"✅ {filename} - 端口配置正确")
                
        except Exception as e:
            print(f"❌ 读取 {filename} 时出错: {e}")
            all_valid = False
    
    return all_valid

def validate_env_files():
    """验证环境配置文件"""
    env_files = ['.env.example', '.env.template']
    all_valid = True
    
    for filename in env_files:
        filepath = Path(filename)
        if not filepath.exists():
            print(f"跳过不存在的环境文件: {filename}")
            continue
            
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
                
            # 检查是否包含必要的端口配置
            required_vars = ['NOFX_BACKEND_PORT', 'NOFX_FRONTEND_PORT', 'API_SERVER_PORT']
            missing_vars = []
            
            for var in required_vars:
                if f'{var}=' not in content:
                    missing_vars.append(var)
                    
            if missing_vars:
                print(f"❌ {filename} 中缺少环境变量: {missing_vars}")
                all_valid = False
            else:
                print(f"✅ {filename} - 环境变量配置完整")
                
        except Exception as e:
            print(f"❌ 读取 {filename} 时出错: {e}")
            all_valid = False
    
    return all_valid

def main():
    print("🔍 开始验证端口配置...")
    print()
    
    print("📋 验证 Docker Compose 文件:")
    compose_valid = validate_docker_compose_ports()
    print()
    
    print("📋 验证环境配置文件:")
    env_valid = validate_env_files()
    print()
    
    if compose_valid and env_valid:
        print("🎉 所有端口配置验证通过！")
        print("💡 现在只需修改 .env 文件中的端口变量，整个项目将统一使用新端口。")
        return True
    else:
        print("❌ 一些配置存在问题，请检查上述错误信息。")
        return False

if __name__ == "__main__":
    main()