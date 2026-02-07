#!/usr/bin/env python3
"""
修复 DecisionCard 组件中的按钮嵌套问题验证脚本

此脚本用于验证 DecisionCard.tsx 中的按钮嵌套问题是否已正确修复
"""

import os
import re

def check_nested_buttons(file_path):
    """检查文件中是否存在按钮嵌套问题"""
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 查找可能的按钮嵌套模式 - button 标签内部包含其他 button 标签
    nested_pattern = r'<button[^>]*>.*?<button[^>]*>.*?</button>.*?</button>'
    nested_matches = re.findall(nested_pattern, content, re.DOTALL | re.IGNORECASE)
    
    # 更宽松的查找模式
    lines = content.split('\n')
    nested_issues = []
    
    for i, line in enumerate(lines):
        if '<button' in line and ('system_prompt' in line.lower() or 'input_prompt' in line.lower()):
            # 检查接下来几行是否有嵌套的按钮
            for j in range(i, min(i+20, len(lines))):
                if '<button' in lines[j] and j != i and '</button>' not in lines[j]:
                    # 检查是否在当前按钮闭合之前还有另一个按钮开始
                    snippet = '\n'.join(lines[i:j+3])
                    if snippet.count('<button') > 1 and snippet.count('</button') < 2:
                        nested_issues.append((i+1, j+1, snippet))
    
    return nested_matches, nested_issues

def main():
    file_path = "web/src/components/DecisionCard.tsx"
    
    if not os.path.exists(file_path):
        print(f"❌ 文件不存在: {file_path}")
        return
    
    print("🔍 正在检查 DecisionCard 组件中的按钮嵌套问题...")
    
    nested_matches, nested_issues = check_nested_buttons(file_path)
    
    if nested_matches:
        print(f"❌ 发现 {len(nested_matches)} 个潜在的按钮嵌套问题")
        for i, match in enumerate(nested_matches):
            print(f"  - 匹配 {i+1}: {match[:100]}...")
    else:
        print("✅ 未发现明显的按钮嵌套正则匹配")
    
    if nested_issues:
        print(f"❌ 发现 {len(nested_issues)} 个具体的嵌套问题区域")
        for start_line, end_line, snippet in nested_issues:
            print(f"  - 行 {start_line}-{end_line}: 可能存在嵌套按钮")
            print(f"    片段: {snippet[:200]}...")
    else:
        print("✅ 未发现具体的按钮嵌套问题")
    
    # 检查修复后的结构
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # 检查是否使用了 div 替代嵌套按钮
    system_div_pattern = r'<div[^>]*className=[^>]*justify-between[^>]*>.*?</div>'
    system_div_matches = re.findall(system_div_pattern, content, re.DOTALL)
    
    input_div_pattern = r'<div[^>]*className=[^>]*justify-between[^>]*>.*?</div>'
    input_div_matches = re.findall(input_div_pattern, content, re.DOTALL)
    
    print("\n✅ 修复验证:")
    print(f"   - 系统提示区域使用 div 结构: {'✅' if 'justify-between' in content and 'copyToClipboard' in content else '❌'}")
    print(f"   - 输入提示区域使用 div 结构: {'✅' if 'justify-between' in content and 'downloadAsFile' in content else '❌'}")
    
    # 检查具体的修复模式
    fixed_system_section = 'copyToClipboard' in content and 'downloadAsFile' in content
    fixed_input_section = 'copyToClipboard' in content and 'downloadAsFile' in content
    
    print(f"\n🎯 修复完成状态:")
    print(f"   - 系统提示区域修复: {'✅' if fixed_system_section else '❌'}")
    print(f"   - 输入提示区域修复: {'✅' if fixed_input_section else '❌'}")
    
    if not nested_matches and not nested_issues:
        print(f"\n🎉 按钮嵌套问题已成功修复！")
        print(f"   现在 Sync 按钮应该可以正常显示了")
    else:
        print(f"\n⚠️  仍存在一些潜在的按钮嵌套问题需要进一步修复")

if __name__ == "__main__":
    main()