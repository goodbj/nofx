#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
更新guardian_client.go文件中的waitForAndValidateOutput函数
"""

import re

def update_function():
    # 读取文件
    with open('mcp/guardian_client.go', 'r', encoding='utf-8') as f:
        content = f.read()

    # 定义新的函数
    new_function = '''func (gc *GuardianClient) waitForAndValidateOutput(ctx context.Context) error {
	// 从环境变量获取配置值
	initialWaitSeconds := getEnvInt("GUARDIAN_INITIAL_WAIT_SECONDS", 60)
	minPollSeconds := getEnvInt("GUARDIAN_MIN_POLL_SECONDS", 2)
	maxPollSeconds := getEnvInt("GUARDIAN_MAX_POLL_SECONDS", 5)
	
	// 前 initialWaitSeconds 秒等待AI输入完成，不做检测
	gc.logger.Printf("⏳ Waiting %%d seconds for AI to start output...", initialWaitSeconds)
	time.Sleep(time.Duration(initialWaitSeconds) * time.Second)
	
	gc.logger.Println("🔍 Starting to poll for AI completion signal...")
	
	// 开始轮询检测，直到找到 ds-flex _0a3d93b 字符串
	for {
		select {
		case <-ctx.Done():
			gc.logger.Println("⏰ Context cancelled, stopping polling for AI completion")
			return ctx.Err()
		default:
			// 获取当前页面的HTML内容
			var pageHTML string
			err := chromedp.Run(ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					err := chromedp.OuterHTML("html", &pageHTML).Do(ctx)
					return err
				}),
			)
			if err != nil {
				gc.logger.Printf("⚠️ Error getting page HTML: %%v", err)
				// 继续尝试，不中断轮询
			} else {
				// 检查是否包含 ds-flex _0a3d93b 字符串
				if strings.Contains(pageHTML, "ds-flex _0a3d93b") {
					gc.logger.Println("✅ AI completion detected: 'ds-flex _0a3d93b' found in page content")
					return nil // 找到关键字，返回成功
				}
			}
			
			// 随机等待 minPollSeconds 到 maxPollSeconds 秒
			randWait := time.Duration(minPollSeconds + rand.Intn(maxPollSeconds-minPollSeconds+1)) * time.Second
			gc.logger.Printf("⏳ Polling wait: checking again in %%0.f seconds", randWait.Seconds())
			time.Sleep(randWait)
		}
	}
}'''

    # 查找并替换函数
    # 首先找到函数的开始位置
    pattern = r'func \(gc \*GuardianClient\) waitForAndValidateOutput\(ctx context\.Context\) error \{.*?\n\}'
    
    # 使用更精确的正则表达式匹配整个函数
    # 这里需要找到完整的函数体，包括所有的花括号
    start_marker = 'func (gc *GuardianClient) waitForAndValidateOutput(ctx context.Context) error {'
    
    if start_marker in content:
        # 找到函数开始位置
        start_pos = content.find(start_marker)
        
        # 从开始位置向后找到函数的结束位置（平衡花括号）
        brace_count = 0
        pos = start_pos
        while pos < len(content):
            if content[pos] == '{':
                brace_count += 1
            elif content[pos] == '}':
                brace_count -= 1
                if brace_count == 0:
                    # 找到匹配的结束括号
                    end_pos = pos + 1
                    break
            pos += 1
        
        # 替换函数
        old_function = content[start_pos:end_pos]
        updated_content = content.replace(old_function, new_function)
        
        # 写回文件
        with open('mcp/guardian_client.go', 'w', encoding='utf-8') as f:
            f.write(updated_content)
            
        print("Function successfully updated!")
    else:
        print("Function not found!")

if __name__ == "__main__":
    update_function()