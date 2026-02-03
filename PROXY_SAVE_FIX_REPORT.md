## 🔧 交易员代理设置保存功能修复报告

### 🐛 问题描述
用户反馈：在交易员编辑页面选择代理设置后，保存没有效果，仍然是原来的设置。

### 🔍 问题分析
经过深入排查，发现两个关键问题：

#### 1. 前端问题 (web/src/components/AITradersPage.tsx)
- `handleSaveEditTrader` 函数在构建请求对象时，遗漏了 `data_access_method` 字段
- 虽然表单数据中包含了该字段，但在发送API请求时没有包含

#### 2. 后端问题 (store/trader.go)
- `Update` 函数在构建数据库更新映射时，遗漏了 `data_access_method` 字段
- 即使前端发送了该字段，也不会被保存到数据库中

### ✅ 修复方案
1. **前端修复**：在 `handleSaveEditTrader` 函数中添加 `data_access_method` 字段到请求对象
2. **后端修复**：在 `Update` 函数中添加 `data_access_method` 字段到数据库更新操作

### 📝 修复详情

#### 文件 1: web/src/components/AITradersPage.tsx
```javascript
// 修复前
const request = {
  name: data.name,
  ai_model_id: data.ai_model_id,
  exchange_id: data.exchange_id,
  strategy_id: data.strategy_id,
  initial_balance: data.initial_balance,
  scan_interval_minutes: data.scan_interval_minutes,
  is_cross_margin: data.is_cross_margin,
  show_in_competition: data.show_in_competition,
  // 缺少 data_access_method 字段
}

// 修复后
const request = {
  name: data.name,
  ai_model_id: data.ai_model_id,
  exchange_id: data.exchange_id,
  strategy_id: data.strategy_id,
  initial_balance: data.initial_balance,
  scan_interval_minutes: data.scan_interval_minutes,
  is_cross_margin: data.is_cross_margin,
  show_in_competition: data.show_in_competition,
  data_access_method: data.data_access_method,  // ✅ 添加此行
}
```

#### 文件 2: store/trader.go
```go
// 修复前
updates := map[string]interface{}{
    "name":                trader.Name,
    "ai_model_id":         trader.AIModelID,
    "exchange_id":         trader.ExchangeID,
    "strategy_id":         trader.StrategyID,
    "is_cross_margin":     trader.IsCrossMargin,
    "show_in_competition": trader.ShowInCompetition,
    // 缺少 data_access_method 字段
}

// 修复后
updates := map[string]interface{}{
    "name":                trader.Name,
    "ai_model_id":         trader.AIModelID,
    "exchange_id":         trader.ExchangeID,
    "strategy_id":         trader.StrategyID,
    "is_cross_margin":     trader.IsCrossMargin,
    "show_in_competition": trader.ShowInCompetition,
    "data_access_method":  trader.DataAccessMethod,  // ✅ 添加此行
}
```

### 🧪 验证步骤
1. 重启后端服务以应用代码更改
2. 在交易员编辑页面修改代理设置（"原生获取" ↔ "代理获取"）
3. 保存后刷新页面，确认设置已正确保存
4. 创建新的交易员时，确认代理设置也能正确保存

### 🎯 影响范围
- 交易员创建和编辑功能
- 代理设置的持久化存储
- 交易员运行时的代理行为（根据配置决定是否使用代理）

### 🚀 部署说明
1. 停止当前运行的后端服务
2. 重新启动后端服务
3. 前端无需额外操作，将自动使用更新后的API

此修复确保了用户在交易员配置中选择的代理设置能够正确保存和应用，解决了用户反馈的问题。