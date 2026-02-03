#!/usr/bin/env python3
# -*- coding: utf-8 -*-

with open('e:/AI/nofx_Dev/mcp/guardian_client.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# 精确修复错误：从第1286行开始到大约1306行是有问题的部分
# 从原始文件我们知道应该在哪个位置接续
fixed_lines = lines[:1286]  # 保留到第1286行

# 添加修复后的代码（替代原来的有问题部分）
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

# 找到应该继续添加后续代码的位置
# 从原文件看，正确的内容应该从大概1307行开始（跳过重复和错误的部分）
# 我们需要跳过错误的重复部分，找到正确的continuation
skip_until_line = 1307  # 跳过有问题的重复部分

# 添加剩余的正确行
fixed_lines.extend(lines[skip_until_line:])

with open('e:/AI/nofx_Dev/mcp/guardian_client_fixed2.go', 'w', encoding='utf-8') as f:
    f.writelines(fixed_lines)

print("Second fixed file created: e:/AI/nofx_Dev/mcp/guardian_client_fixed2.go")