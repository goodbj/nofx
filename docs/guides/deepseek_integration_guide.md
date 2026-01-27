# DeepSeek 集成指南

## 概述

本指南介绍如何在 NoFx 系统中正确配置和使用 DeepSeek AI 服务。系统支持两种 DeepSeek 集成方式：

1. **API 方式**（推荐）：通过 `https://api.deepseek.com` 直接调用 API
2. **Web 浏览器自动化方式**：通过 `https://chat.deepseek.com` 使用浏览器自动化（Guardian）

## 推荐配置：API 方式

### 1. 获取 DeepSeek API Key

1. 访问 [DeepSeek API Keys 页面](https://platform.deepseek.com/api_keys)
2. 登录您的 DeepSeek 账户
3. 创建新的 API Key
4. 记住 API Key，稍后配置时会用到

### 2. 配置环境变量

在您的 `.env` 文件中添加：

```bash
DEEPSEEK_API_KEY=sk-your-deepseek-api-key-here
```

### 3. 在交易员配置中使用

当创建或编辑交易员时，选择 "DeepSeek" 作为 AI 模型，并确保在高级设置中使用 API 方式。

## Web 浏览器自动化方式（Guardian）

### 何时使用

Web 浏览器自动化方式适用于以下场景：

- 您没有 DeepSeek API Key
- 您想利用 DeepSeek 网页版的某些特殊功能
- API 无法访问时的备用方案

### 配置步骤

1. 确保系统已安装 Chrome 或 Chromium 浏览器
2. 在环境变量中设置：

```bash
GUARDIAN_AI_PROVIDER=deepseek
GUARDIAN_AI_ENDPOINT=https://chat.deepseek.com
GUARDIAN_PAGE_URL=https://chat.deepseek.com
```

### 注意事项

Web 浏览器自动化方式可能存在以下问题：

- **页面结构变化**：DeepSeek 可能会更新其网页结构，导致选择器失效
- **反爬虫机制**：可能触发反爬虫保护
- **登录要求**：可能需要登录才能使用
- **性能影响**：相比 API 方式，性能较低且不稳定

## 故障排除

### Web 端获取信息错误

如果您遇到 "Web 端获取信息错误" 的问题，请尝试以下解决方案：

#### 1. 优先使用 API 方式

最可靠的解决方案是切换到 API 方式：

1. 获取 DeepSeek API Key
2. 在配置中选择 "DeepSeek" API 方式而不是 "Guardian" 方式

#### 2. 检查浏览器自动化配置

如果必须使用 Web 方式，请检查：

1. **浏览器安装**：确保 Chrome/Chromium 已正确安装
2. **网络连接**：确保可以访问 `https://chat.deepseek.com`
3. **页面元素**：DeepSeek 可能已更新页面结构，需要更新选择器

#### 3. 日志诊断

启用详细日志以诊断问题：

```bash
LOG_LEVEL=debug
GUARDIAN_DEBUG=true
```

## 最佳实践

1. **优先使用 API 方式**：API 方式更稳定、更快、更可靠
2. **备份方案**：可以同时配置 API 和 Web 方式作为备份
3. **定期更新**：如果使用 Web 方式，定期检查页面结构是否变化
4. **监控性能**：关注 API 调用的响应时间和成功率

## 环境变量配置

| 变量 | 说明 | 示例 |
|------|------|------|
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 | `sk-xxx...` |
| `GUARDIAN_AI_PROVIDER` | Guardian AI 提供商 | `deepseek` |
| `GUARDIAN_AI_ENDPOINT` | Guardian AI 端点 | `https://chat.deepseek.com` |
| `GUARDIAN_PAGE_URL` | Guardian 页面 URL | `https://chat.deepseek.com` |
| `GUARDIAN_DISPLAY_ENABLED` | 是否启用显示 | `false` (Docker 中建议禁用) |

## 支持

如遇到问题，请：

1. 检查是否使用了推荐的 API 方式
2. 查看日志中的详细错误信息
3. 确认网络连接和 API Key 有效性
4. 参考完整的错误日志进行诊断