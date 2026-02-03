#!/usr/bin/env python3
# -*- coding: utf-8 -*-

with open('e:/AI/nofx_Dev/mcp/guardian_client.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# 修复问题 - 替换有问题的部分
# 我们需要修复第1286-1306行之间的重复声明和未定义变量问题
fixed_lines = lines[:1286]  # 保留到第1286行

# 添加修复后的代码，去除重复的pageHTML声明，并正确调用extractAIOutputContent
fixed_lines.append('\t\t\t\t// 检查页面是否包含 ds-flex _0a3d93b 字符串\n')
fixed_lines.append('\t\t\t\tif err == nil && strings.Contains(pageHTML, "ds-flex _0a3d93b") {\n')
fixed_lines.append('\t\t\t\t\t// 找到关键词，提取内容\n')
fixed_lines.append('\t\t\t\t\textractedContent, extractErr := gc.extractAIOutputContent(ctx)\n')
fixed_lines.append('\t\t\t\t\tif extractErr != nil {\n')
fixed_lines.append('\t\t\t\t\t\tgc.logger.Printf("⚠️ Error extracting AI output content: %v", extractErr)\n')
fixed_lines.append('\t\t\t\t\t\t// 如果提取失败，仍然返回成功，但内容为空\n')
fixed_lines.append('\t\t\t\t\t\treturn "", nil\n')
fixed_lines.append('\t\t\t\t\t}\n')
fixed_lines.append('\t\t\t\t\tgc.logger.Printf("✅ Successfully extracted AI output content, length: %d", len(extractedContent))\n')
fixed_lines.append('\t\t\t\t\treturn extractedContent, nil // 找到关键字并提取内容，返回成功\n')
fixed_lines.append('\t\t\t\t}\n')
fixed_lines.append('\t\t\t}\n')

# 添加剩余的行
fixed_lines.extend(lines[1298:])

with open('e:/AI/nofx_Dev/mcp/guardian_client_fixed.go', 'w', encoding='utf-8') as f:
    f.writelines(fixed_lines)

print("Fixed file created: e:/AI/nofx_Dev/mcp/guardian_client_fixed.go")