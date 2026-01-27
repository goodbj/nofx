# Guardian 浏览器自动化功能配置指南

## 概述

Guardian 浏览器自动化功能允许系统将提示词发送到浏览器中的AI服务（如DeepSeek、ChatGPT、Claude等），然后获取AI的响应。本文档介绍如何配置和使用此功能。

## 支持的AI服务

目前支持以下AI服务：

- **DeepSeek** - 默认配置
- **ChatGPT** - OpenAI ChatGPT
- **Claude** - Anthropic Claude
- **Gemini** - Google Gemini

## 配置方式

### 1. 环境变量配置

通过环境变量指定AI服务类型：

```bash
# 设置为DeepSeek（默认）
export GUARDIAN_AI_SERVICE=deepseek

# 设置为ChatGPT
export GUARDIAN_AI_SERVICE=chatgpt

# 设置为Claude
export GUARDIAN_AI_SERVICE=claude

# 设置为Gemini
export GUARDIAN_AI_SERVICE=gemini
```

### 2. Docker环境配置

在Docker环境中，可以在docker-compose文件中添加环境变量：

```yaml
services:
  backend:
    environment:
      - GUARDIAN_AI_SERVICE=deepseek  # 或其他支持的服务
```

### 3. 选择器配置

每个AI服务都有不同的DOM结构，Guardian使用以下预定义的选择器：

#### DeepSeek
- **输入框**: `#prompt-textarea`
- **提交按钮**: `button[type='submit']`
- **响应区域**: `.font-light`

#### ChatGPT
- **输入框**: `textarea[placeholder*='Send a message']`
- **提交按钮**: `button[data-testid='send-button']`
- **响应区域**: `[data-message-author-role='assistant']`

#### Claude
- **输入框**: `div[data-is-empty='true'] div`
- **提交按钮**: `button[data-testid='send-button']`
- **响应区域**: `div[data-testid='assistant-reply'] div`

#### Gemini
- **输入框**: `textarea[aria-label*='Describe what you need']`
- **提交按钮**: `button[aria-label*='Send']`
- **响应区域**: `div[data-read-aloud]`

## 使用方法

### 1. 前端配置

在交易员配置页面中：

1. 将AI模型类型设置为 `guardian`
2. 在Base URL字段中输入目标AI服务的URL（可选，如果不设置则使用默认的DeepSeek服务）
3. 保存配置

**支持的目标AI服务**：
- https://chat.deepseek.com/ (DeepSeek)
- https://chat.openai.com/ (ChatGPT)
- https://claude.ai/ (Claude)
- https://gemini.google.com/ (Gemini)
- https://www.bing.com/chat (Bing Chat)
- https://www.perplexity.ai/ (Perplexity)

系统会根据URL自动识别并使用相应的DOM选择器配置。

### 2. 系统行为

当AI模型设置为"guardian"时：

1. 系统会自动使用GuardianClient
2. 提示词会被发送到指定的AI服务
3. 从浏览器中获取AI响应
4. 响应会被处理并用于后续决策

## 注意事项

1. **Chrome浏览器**：确保系统上安装了Chrome或Chromium浏览器
2. **网络连接**：确保能够访问目标AI服务
3. **登录状态**：某些AI服务需要登录才能使用，可能需要配置用户数据目录
4. **选择器适配**：AI服务的UI可能会更新，需要相应调整选择器
5. **性能考虑**：浏览器自动化比API调用慢，请合理安排调用频率

## 故障排除

### 常见问题

1. **浏览器无法启动**
   - 检查系统是否安装了Chrome/Chromium
   - 检查权限设置

2. **选择器不匹配**
   - 检查目标AI服务的UI是否有更新
   - 验证选择器配置是否正确

3. **响应获取失败**
   - 检查AI服务是否正常工作
   - 确认是否有登录要求

### 调试技巧

1. 启用详细日志输出
2. 使用非静默模式（headless=false）观察浏览器操作
3. 检查AI服务的开发者工具控制台

## 高级配置

如需自定义选择器，可通过修改源代码中的选择器配置来适配特定的AI服务或UI更新。