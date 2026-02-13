# Dashboard页面"周期"内容数据源分析报告

## 🔍 数据存储位置确认

**Dashboard页面的"周期"内容存储在以下位置：**

### 📊 数据库信息
- **数据库文件**: `data/data.db` (SQLite数据库)
- **数据表名**: `decision_records`
- **记录总数**: 2,338条记录
- **涉及交易员**: 14个不同的交易员
- **最大周期号**: 1641

### 🏗️ 表结构详情
`decision_records`表包含以下关键字段：
- `id` (INTEGER) - 主键
- `trader_id` (TEXT) - 交易员ID
- `cycle_number` (INTEGER) - 周期号（显示在页面上的"周期"编号）
- `timestamp` (datetime) - 时间戳
- `success` (numeric) - 成功状态（1=成功, 0=失败）
- `error_message` (TEXT) - 错误信息
- `system_prompt` (TEXT) - 系统提示词
- `input_prompt` (TEXT) - 输入提示词
- `decision_json` (TEXT) - 决策JSON
- `candidate_coins` (TEXT) - 候选币种
- `execution_log` (TEXT) - 执行日志

## 🔄 数据流向

### 1. 数据生成流程
```
交易员AI决策 → DecisionStore.LogDecision() → 写入decision_records表
```

### 2. 数据获取流程
```
decision_records表 → 后端API(/api/decisions/latest) → 前端useSWR → DecisionCard组件显示
```

### 3. 具体实现路径

**后端 (Go):**
- API路由: `/api/decisions/latest`
- 处理函数: `handleLatestDecisions()` in `server.go`
- 数据访问: `DecisionStore.GetLatestRecords()` in `store/decision.go`
- 查询逻辑: 按`timestamp DESC`排序，限制返回条数

**前端 (React/TypeScript):**
- API调用: `api.getLatestDecisions()` in `web/src/lib/api.ts`
- 数据获取: `useSWR` hook with 30秒刷新间隔
- 组件显示: `DecisionCard` component in `web/src/components/DecisionCard.tsx`
- 页面集成: `TraderDashboardPage` in `web/src/pages/TraderDashboardPage.tsx`

## 📈 当前数据状态

### 运行中交易员数据
- **交易员名称**: 虚拟盘通用交易员
- **交易员ID**: bac26dfa_guardian-ai_1770901210
- **该交易员记录数**: 32条
- **最近记录**:
  - Cycle 40 | 2026-02-13 02:31:30 | 成功
  - Cycle 39 | 2026-02-13 02:27:48 | 成功
  - Cycle 38 | 2026-02-13 02:22:06 | 成功

### 数据统计
- **总记录数**: 2,338条
- **成功记录**: 1,122条 (48%)
- **失败记录**: 1,216条 (52%)
- **涉及交易员**: 14个

## 🔧 技术实现细节

### 数据库查询示例
```sql
SELECT 
    cycle_number,
    timestamp,
    success,
    error_message
FROM decision_records 
WHERE trader_id = 'bac26dfa_guardian-ai_1770901210'
ORDER BY timestamp DESC 
LIMIT 5
```

### API调用示例
```javascript
// 前端调用
const { data: decisions } = useSWR<DecisionRecord[]>(
    `decisions/latest-${selectedTraderId}-${decisionsLimit}`,
    () => api.getLatestDecisions(selectedTraderId, decisionsLimit),
    {
        refreshInterval: 30000, // 30秒刷新
        revalidateOnFocus: false,
    }
)
```

### 后端处理逻辑
```go
// server.go中的处理函数
func (s *Server) handleLatestDecisions(c *gin.Context) {
    // 获取交易员ID和限制数量
    // 调用DecisionStore.GetLatestRecords()
    // 返回JSON格式的决策记录
}
```

## ✅ 结论

Dashboard页面显示的"周期"内容完全来自于`data/data.db`数据库中的`decision_records`表，通过标准的API接口和前端组件进行展示。数据更新正常，系统架构清晰，数据流向完整。

**关键要点：**
1. 数据存储在本地SQLite数据库中
2. 通过RESTful API接口提供数据服务
3. 前端使用SWR进行数据获取和缓存
4. 数据刷新间隔为30秒
5. 系统运行正常，数据完整