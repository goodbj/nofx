# 手动AI工作流使用指南 (Manual AI Workflow Guide)

> **免费使用DeepSeek进行交易决策，无需付费API，无需本地部署大模型**

## 📌 功能概述

手动AI工作流允许您通过免费的DeepSeek Web版本（chat.deepseek.com）进行AI交易决策，然后将决策结果提交给NoFx执行实际交易。

### 优势
- ✅ **完全免费** - 使用DeepSeek免费网页版
- ✅ **无需API密钥** - 不需要付费的AI API服务
- ✅ **无需本地部署** - 不需要搭建本地大模型服务器
- ✅ **实时数据** - 使用真实的市场数据和持仓信息
- ✅ **完整决策链** - 从Prompt生成到交易执行的完整流程

---

## 🚀 使用流程

### 步骤1: 生成完整Prompt

1. 登录NoFx系统
2. 前往 **"策略工作室"** 页面
3. 找到您要测试的策略配置
4. 点击 **"AI测试"** 按钮
5. 选择一个交易员
6. 系统会生成包含以下内容的完整Prompt：
   - System Prompt（策略规则和决策要求）
   - User Prompt（实时市场数据、持仓信息、技术指标等）

### 步骤2: 复制Prompt到DeepSeek

1. 在AI测试窗口中，点击 **"复制完整Prompt (System + User)"** 按钮
2. 打开浏览器访问 [chat.deepseek.com](https://chat.deepseek.com)
3. 将复制的Prompt粘贴到DeepSeek对话框中
4. 发送给DeepSeek

### 步骤3: 获取AI决策

1. 等待DeepSeek生成决策（通常几秒钟）
2. DeepSeek会返回一个JSON格式的决策数组
3. 复制完整的JSON决策（包括方括号`[]`）

**决策JSON示例：**
```json
[
  {
    "symbol": "BTCUSDT",
    "action": "open_long",
    "leverage": 5,
    "position_size_usd": 100,
    "stop_loss": 90000,
    "take_profit": 105000,
    "confidence": 85,
    "reasoning": "技术分析显示强势突破，RSI超卖反弹"
  }
]
```

### 步骤4: 提交决策到NoFx

1. 前往 **"测试代码"** 页面
2. 切换到 **"AI Prompt 查看"** 标签
3. 滚动到页面底部找到 **"免费AI工作流（DeepSeek Web版）"** 区域
4. 在 **"选择交易员"** 下拉框中选择要执行决策的交易员
5. 将从DeepSeek复制的JSON粘贴到 **"粘贴DeepSeek生成的决策JSON"** 文本框
6. 点击 **"🚀 执行AI决策"** 按钮

### 步骤5: 查看执行结果

系统会显示执行结果，包括：
- 总决策数量
- 成功执行数量
- 失败数量
- 每个决策的详细执行状态

---

## 📋 支持的决策动作 (Action Types)

### 基础开平仓
- `open_long` - 开多单
- `open_short` - 开空单
- `close_long` - 平多单
- `close_short` - 平空单

### 观望动作
- `hold` - 持仓不动
- `wait` - 等待观望

### 高级操作
- `update_stop_loss` - 修改止损
- `update_take_profit` - 修改止盈
- `partial_close` - 部分平仓
- `trailing_stop` - 追踪止损
- `dynamic_take_profit` - 动态止盈

---

## ⚙️ 技术实现细节

### 后端API端点

#### 1. 生成完整Prompt (已简化)
```
POST /api/test/generate-full-prompt
```
**说明**：返回用户指引，建议使用策略页面的"AI测试"功能

#### 2. 提交AI决策
```
POST /api/test/submit-ai-decision
```

**请求参数**：
```json
{
  "trader_id": "trader_id_here",
  "decision_json": "[{...}]"  // JSON字符串
}
```

**响应示例**：
```json
{
  "success": true,
  "total": 1,
  "success_count": 1,
  "fail_count": 0,
  "results": [
    {
      "index": 1,
      "symbol": "BTCUSDT",
      "action": "open_long",
      "success": true,
      "message": "Executed successfully"
    }
  ],
  "message": "Executed 1/1 decisions successfully"
}
```

---

## 🔒 安全注意事项

1. **验证交易员权限** - 系统会自动验证您是否有权限操作选定的交易员
2. **JSON格式验证** - 确保从DeepSeek复制的是有效的JSON格式
3. **决策审核** - 在执行前请仔细检查AI生成的决策是否合理
4. **风险控制** - 建议先使用测试网交易所进行验证

---

## 🐛 常见问题

### Q1: 为什么不直接在NoFx中生成完整Prompt？
**A**: 直接生成需要访问trader的复杂内部状态（账户、持仓、市场数据等），实现复杂且容易出错。使用现有的"AI测试"功能更稳定可靠。

### Q2: 可以使用其他AI服务吗（如ChatGPT、Claude）？
**A**: 可以！只要能理解JSON格式并返回符合规范的决策即可。但DeepSeek在中文和代码理解方面表现优秀且免费。

### Q3: 执行失败怎么办？
**A**: 查看执行结果中的详细错误信息，常见原因包括：
- JSON格式错误
- 缺少必需参数（如`leverage`、`position_size_usd`）
- 交易员未启动或连接失败
- 交易所API限制

### Q4: 可以自动化这个流程吗？
**A**: 可以！未来可以开发守护程序（Phase 2），使用浏览器自动化（Selenium/Playwright）实现：
1. 自动将Prompt粘贴到DeepSeek
2. 等待AI生成决策
3. 自动复制结果并提交给NoFx

---

## 📚 相关文件

### 后端实现
- `/api/server.go` (L227-228) - 路由注册
- `/api/strategy.go` (L815-949) - API处理函数
  - `handleGenerateFullPrompt` - 生成Prompt（简化版）
  - `handleSubmitAIDecision` - 提交决策执行

### 前端实现
- `/web/src/pages/TestPage.tsx` - 测试页面UI
  - 手动AI工作流区域 (L1172-1281)
  - 提交决策逻辑 (L628-672)

### 文档
- `/kernel/engine.go` (L1057-1064) - Action类型文档
- `/full_sample_prompt.txt` (L135-143) - 示例Prompt

---

## 🎯 未来扩展（Phase 2）

### 自动化守护程序
通过Python + Selenium/Playwright实现：
```python
# 伪代码示例
while True:
    prompt = nofx_api.get_full_prompt(trader_id)
    
    # 浏览器自动化
    deepseek.paste_prompt(prompt)
    decision_json = deepseek.wait_and_copy_response()
    
    # 提交执行
    result = nofx_api.submit_decision(trader_id, decision_json)
    
    time.sleep(interval)
```

### 关键挑战
- DeepSeek网页结构变化检测
- 验证码处理
- 会话保持
- 错误恢复机制

---

## 📞 技术支持

如有问题，请查看：
- [NoFx完整文档](../README.md)
- [API文档](../api/)
- [社区讨论](../community/)

**祝交易顺利！** 🚀
