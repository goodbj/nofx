# 劫匪系统（浏览器自动化）开发文档

## 1. 系统概述

劫匪系统是 NoFx 平台的浏览器自动化功能模块，用于在无法直接使用 API 的情况下，通过模拟浏览器操作与 AI 服务进行交互。该系统允许用户绕过 API 限制，直接与各种 AI 服务的网页版进行交互。

## 2. 核心组件

### 2.1 GuardianClient（守护者客户端）
- **位置**: `mcp/guardian_client.go`
- **作用**: 浏览器自动化的核心实现类
- **功能**:
  - 启动和控制 Chrome 浏览器实例
  - 导航到目标 AI 服务网站
  - 自动填写提示词到输入框
  - 点击提交按钮
  - 等待并提取 AI 响应内容

### 2.2 BypassClient（旁路客户端）
- **位置**: `mcp/bypass_client.go`
- **作用**: 作为中间层，拦截标准 API 调用并将其重定向到浏览器自动化
- **功能**:
  - 检测是否启用了浏览器自动化模式
  - 根据配置决定使用标准 API 或浏览器自动化
  - 透明地将请求转发到相应实现

### 2.3 AIModel 存储
- **位置**: `store/ai_model.go`
- **作用**: 存储 AI 模型配置，包括 `useBrowserAutomation` 标志
- **功能**:
  - 保存用户的 AI 模型配置
  - 提供 `useBrowserAutomation` 字段来控制是否使用浏览器自动化

## 3. 工作流程

### 3.1 启用劫匪系统的流程
1. 用户在 AI 模型配置中勾选"绕过API - 使用浏览器自动化"复选框
2. 系统将 `useBrowserAutomation` 标志设置为 `true` 并保存到数据库
3. 当交易员使用该 AI 模型时，系统创建 `BypassClient` 实例
4. `BypassClient` 检测到 `useBrowserAutomation` 为 `true`，使用 `GuardianClient` 处理请求

### 3.2 浏览器自动化执行流程
1. `GuardianClient` 启动 Chrome 浏览器实例（非 headless 模式）
2. 导航到指定的 AI 服务网站（如 https://chat.deepseek.com）
3. 使用预定义的选择器定位输入框
4. 填入提示词内容
5. 点击提交按钮
6. 等待 AI 生成响应
7. 提取并返回 AI 响应内容

## 4. 配置与控制

### 4.1 数据库字段
- `use_browser_automation` (布尔型): 控制是否启用浏览器自动化
- 默认值: `false`

### 4.2 前端界面
- 在 AI 模型配置界面提供"绕过API - 使用浏览器自动化"复选框
- 在交易员配置界面同样提供该选项（但已移除）

### 4.3 API 端点
- `GET /api/models`: 获取 AI 模型配置（包含 `useBrowserAutomation` 字段）
- `PUT /api/models`: 更新 AI 模型配置（包含 `use_browser_automation` 参数）

## 5. 支持的 AI 服务

劫匪系统支持以下 AI 服务：
- DeepSeek (`deepseek`)
- OpenAI/ChatGPT (`openai`)
- Anthropic Claude (`claude`)
- Google Gemini (`gemini`)
- Alibaba Qwen (`qwen`)
- xAI Grok (`grok`)
- Moonshot Kimi (`kimi`)
- Ollama (`ollama`)
- 以及其他自定义服务

## 6. 选择器策略

系统使用多层次选择器策略来定位页面元素：

### 6.1 输入框选择器（按优先级排序）
```
- textarea[placeholder='给 DeepSeek 发送消息 ']
- textarea._27c9245.ds-scroll-area.d96f2d2a
- textarea[placeholder='Send a message']
- div.public-DraftEditor-content
- textarea[aria-label='Chat text input']
```

### 6.2 提交按钮选择器
```
- div._7436101.ds-icon-button[role='button']:not([aria-disabled='true'])
- button[type='submit']
- button[data-testid='send-button']
```

### 6.3 响应内容选择器
```
- div.ds-flex._0a3d93b
- div.text-message span
- div.markdown-body
- pre, code
```

## 7. 安全与隐私

### 7.1 数据加密
- 所有敏感配置信息（如 API 密钥）均使用加密存储
- 传输过程中的数据也经过加密

### 7.2 浏览器安全
- 使用用户数据目录保存登录状态
- 配置防自动化检测选项
- 设置合适的用户代理字符串

## 8. 故障排除

### 8.1 常见问题
1. **选择器失效**: AI 服务更新页面结构时需要更新选择器
2. **反爬虫机制**: 某些网站可能检测并阻止自动化操作
3. **超时问题**: 设置合适的超时时间避免长时间等待

### 8.2 调试方法
- 启用非 headless 模式观察浏览器操作
- 检查日志输出了解执行流程
- 测试不同选择器的有效性

## 9. 维护与扩展

### 9.1 添加新的 AI 服务支持
1. 在 `GuardianClient` 中添加服务特定的配置
2. 定义新的选择器集合
3. 实现特定于服务的交互逻辑

### 9.2 优化建议
- 定期更新选择器以适应页面结构变化
- 实现智能重试机制
- 添加更多容错处理

## 10. 架构优势

1. **统一接口**: 所有 AI 服务通过统一的浏览器自动化接口
2. **无缝切换**: 用户可以随时在 API 模式和浏览器自动化模式之间切换
3. **透明实现**: 上层代码无需关心底层实现差异
4. **可扩展性**: 易于添加新的 AI 服务支持