# NOFX API接口详细文档

## API概述

NOFX API采用RESTful设计风格，提供完整的交易系统管理功能。所有API端点均以 `/api/v1/` 为前缀。

### 认证机制
- **JWT Token**: 基于Bearer Token的身份验证
- **Token有效期**: 默认7天（可配置）
- **刷新机制**: 支持Token续期

### 响应格式
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2026-03-02T10:30:00Z"
}
```

### 错误响应格式
```json
{
  "success": false,
  "error": "错误代码",
  "message": "错误描述",
  "details": {}
}
```

## 认证相关接口

### 用户登录
```
POST /api/v1/auth/login
```

**请求参数**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_string",
    "user": {
      "id": "user_id",
      "username": "username",
      "email": "user@example.com"
    }
  }
}
```

### 用户注册
```
POST /api/v1/auth/register
```

**请求参数**:
```json
{
  "username": "string",
  "email": "string",
  "password": "string",
  "confirm_password": "string"
}
```

### 刷新Token
```
POST /api/v1/auth/refresh
```

**请求头**:
```
Authorization: Bearer <refresh_token>
```

## 交易器管理接口

### 获取交易器列表
```
GET /api/v1/traders
```

**查询参数**:
- `status`: 运行状态 (all/running/stopped)
- `exchange`: 交易所筛选
- `page`: 页码
- `limit`: 每页数量

**响应示例**:
```json
{
  "success": true,
  "data": {
    "traders": [
      {
        "id": "trader_001",
        "name": "BTC-USDT策略",
        "exchange_id": "binance",
        "ai_model_id": "deepseek_chat",
        "strategy_id": "strategy_001",
        "is_running": true,
        "created_at": "2026-03-01T10:00:00Z",
        "updated_at": "2026-03-02T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 10
  }
}
```

### 创建交易器
```
POST /api/v1/traders
```

**请求参数**:
```json
{
  "name": "交易器名称",
  "exchange_id": "交易所ID",
  "ai_model_id": "AI模型ID",
  "strategy_id": "策略ID",
  "config": {
    "trading_pair": "BTCUSDT",
    "leverage": 10,
    "position_size": 0.01
  }
}
```

### 启动交易器
```
POST /api/v1/traders/{id}/start
```

**响应**:
```json
{
  "success": true,
  "message": "交易器启动成功"
}
```

### 停止交易器
```
POST /api/v1/traders/{id}/stop
```

### 删除交易器
```
DELETE /api/v1/traders/{id}
```

## 策略管理接口

### 获取策略列表
```
GET /api/v1/strategies
```

### 创建策略
```
POST /api/v1/strategies
```

**请求参数**:
```json
{
  "name": "策略名称",
  "description": "策略描述",
  "coin_source": {
    "type": "static",
    "coins": ["BTCUSDT", "ETHUSDT"]
  },
  "indicators": {
    "ema_periods": [9, 21, 55],
    "enable_macd": true,
    "enable_rsi": true
  },
  "risk_control": {
    "max_position_size": 0.1,
    "stop_loss_percent": 5.0,
    "take_profit_percent": 10.0
  },
  "ai_prompt": {
    "system_prompt": "你是专业的交易员...",
    "user_prompt_template": "分析以下数据..."
  }
}
```

### 更新策略
```
PUT /api/v1/strategies/{id}
```

### 删除策略
```
DELETE /api/v1/strategies/{id}
```

### 测试策略AI响应
```
POST /api/v1/strategies/{id}/test
```

## 回测接口

### 启动回测
```
POST /api/v1/backtest/run
```

**请求参数**:
```json
{
  "strategy_id": "策略ID",
  "symbols": ["BTCUSDT", "ETHUSDT"],
  "timeframe": "1h",
  "start_date": "2024-01-01",
  "end_date": "2024-12-31",
  "initial_capital": 10000,
  "config": {
    "commission": 0.001,
    "slippage": 0.0005
  }
}
```

### 获取回测列表
```
GET /api/v1/backtest/runs
```

### 获取回测详情
```
GET /api/v1/backtest/runs/{id}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": "backtest_001",
    "strategy_id": "strategy_001",
    "status": "completed",
    "progress": 100,
    "start_time": "2026-03-02T09:00:00Z",
    "end_time": "2026-03-02T09:30:00Z",
    "results": {
      "total_return": 15.5,
      "annual_return": 45.2,
      "sharpe_ratio": 1.8,
      "max_drawdown": 12.3,
      "win_rate": 65.2,
      "total_trades": 156
    },
    "equity_curve": [
      {"timestamp": "2024-01-01T00:00:00Z", "equity": 10000},
      {"timestamp": "2024-01-02T00:00:00Z", "equity": 10150}
    ]
  }
}
```

### 实时回测进度
```
GET /api/v1/backtest/runs/{id}/progress
```
*使用Server-Sent Events (SSE) 实时推送进度*

## AI辩论接口

### 创建辩论会话
```
POST /api/v1/debate/sessions
```

**请求参数**:
```json
{
  "topic": "是否应该买入BTC",
  "symbols": ["BTCUSDT"],
  "max_rounds": 3,
  "models": ["deepseek", "qwen", "gpt-4"],
  "roles": {
    "deepseek": "bull",
    "qwen": "bear",
    "gpt-4": "analyst"
  }
}
```

### 获取辩论列表
```
GET /api/v1/debate/sessions
```

### 获取辩论详情
```
GET /api/v1/debate/sessions/{id}
```

### 实时辩论进度
```
GET /api/v1/debate/sessions/{id}/stream
```
*使用SSE实时推送辩论过程*

## 市场数据接口

### 获取K线数据
```
GET /api/v1/market/kline
```

**查询参数**:
- `symbol`: 交易对 (如: BTCUSDT)
- `interval`: 时间周期 (1m, 5m, 15m, 1h, 4h, 1d)
- `limit`: 数据条数 (默认100, 最大1000)

### 获取实时价格
```
GET /api/v1/market/price/{symbol}
```

### 获取技术指标
```
GET /api/v1/market/indicators
```

**查询参数**:
- `symbol`: 交易对
- `indicators`: 指标列表 (ema,macd,rsi,atr)
- `periods`: 周期参数

## 系统管理接口

### 系统健康检查
```
GET /api/v1/health
```

**响应**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2026-03-02T10:30:00Z",
    "version": "1.0.0"
  }
}
```

### 系统状态
```
GET /api/v1/system/status
```

### 系统指标
```
GET /api/v1/system/metrics
```

### 配置管理
```
GET /api/v1/config/public
PUT /api/v1/config/system
```

## 环境变量接口

### 获取环境变量
```
GET /api/v1/env/variables
```

### 更新环境变量
```
PUT /api/v1/env/variables
```

## 错误代码说明

| 错误代码 | 说明 | HTTP状态码 |
|---------|------|-----------|
| AUTH_001 | 认证失败 | 401 |
| AUTH_002 | Token过期 | 401 |
| AUTH_003 | 权限不足 | 403 |
| VALID_001 | 参数验证失败 | 400 |
| TRADE_001 | 交易器不存在 | 404 |
| TRADE_002 | 交易器已在运行 | 409 |
| TRADE_003 | 交易器启动失败 | 500 |
| STRAT_001 | 策略不存在 | 404 |
| BACK_001 | 回测任务不存在 | 404 |
| BACK_002 | 回测任务进行中 | 409 |
| AI_001 | AI模型调用失败 | 500 |
| DB_001 | 数据库操作失败 | 500 |

## 限流配置

- **API请求限制**: 1000次/分钟/IP
- **认证接口**: 100次/分钟/IP
- **回测接口**: 10次/小时/IP
- **AI接口**: 50次/分钟/IP

## WebSocket支持

### 实时数据推送
```
连接地址: ws://localhost:8888/ws
```

**订阅主题**:
- `market.{symbol}.price` - 实时价格
- `market.{symbol}.kline` - K线数据
- `trader.{id}.status` - 交易器状态
- `backtest.{id}.progress` - 回测进度
- `debate.{id}.update` - 辩论更新

## SDK客户端示例

### Python客户端
```python
import requests

class NOFXClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        })
    
    def get_traders(self):
        response = self.session.get(f'{self.base_url}/api/v1/traders')
        return response.json()
    
    def start_trader(self, trader_id):
        response = self.session.post(f'{self.base_url}/api/v1/traders/{trader_id}/start')
        return response.json()
```

### JavaScript客户端
```javascript
class NOFXClient {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
        this.headers = {
            'Authorization': `Bearer ${apiKey}`,
            'Content-Type': 'application/json'
        };
    }
    
    async getTraders() {
        const response = await fetch(`${this.baseUrl}/api/v1/traders`, {
            headers: this.headers
        });
        return response.json();
    }
}
```

## API版本历史

### v1.0 (当前版本)
- 基础交易器管理
- 策略配置接口
- 回测功能接口
- AI辩论接口
- 系统管理接口

### 未来版本规划
- v1.1: 增加订单管理接口
- v1.2: 增加风险监控接口
- v1.3: 增加数据导出接口
- v2.0: GraphQL接口支持

---
*API文档版本: v1.0*
*最后更新: 2026年3月*