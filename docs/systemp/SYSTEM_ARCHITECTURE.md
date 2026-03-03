# NOFX系统架构文档

## 项目概述

NOFX 是一个基于 AI 的多资产量化交易平台，支持加密货币、美股、外汇和贵金属等多市场交易。系统采用前后端分离架构，使用 Go 语言构建后端服务，React + TypeScript构建前端界面。

##技术栈

###后端技术栈
- **语言**: Go 1.25.3
- **Web框架**: Gin v1.11.0
- **数据库**: SQLite (默认) / PostgreSQL (可选)
- **AI模型**: DeepSeek, Qwen, OpenAI (GPT), Claude, Gemini, Grok, Kimi
- **交易所API**: Binance, Bybit, OKX, Bitget, Hyperliquid, Aster DEX, Lighter
- **技术指标库**: TA-Lib
- **加密**: AES-256 + RSA

###技术栈
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 6.4
- **状态管理**: Zustand
- **UI库**: Ant Design 6.2 + TailwindCSS
- **图表**: Lightweight Charts + Recharts
- **数据获取**: SWR
- **动画**: Framer Motion

##系统架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              NOFX Platform                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐│
│  │  Strategy   │  │  Backtest   │  │   Debate    │  │   Live Trading      ││
│  │   Studio    │  │   Engine    │  │    Arena    │  │   (Auto Trader)     ││
│  └──────┬──────┘  └──────┬──────┘ └──────┬──────┘ └──────────┬──────────┘│
│         │                │                │                    │           │
│         └────────────────┴────────────────┴────────────────────┘           │
│                                    │                                        │
│                          ┌─────────▼─────────┐                              │
│                          │   Core Services   │                              │
│                          │  - Market Data    │                              │
│                          │  - AI Providers   │                              │
│                          │  - Risk Control   │                              │
│                          └─────────┬─────────┘                              │
│                                    │                                        │
│        ┌──────────────────────────┼──────────────────────────┐            │
│         │                          │                          │            │
│  ┌──────▼──────┐         ┌─────────▼─────────┐      ┌────────▼────────┐   │
│  │  Exchanges  │         │     Database      │      │   Frontend UI   │   │
│  │  (CEX/DEX)  │         │    (SQLite)       │      │   (React SPA)   │   │
│  └─────────────┘         └───────────────────┘      └─────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

##核心模块结构

### 1. 主入口模块
```
main.go                    #应用程序入口点
├──配置初始化
├── 数据库连接
├── 加密服务初始化
├── TraderManager 初始化
├── BacktestManager 初始化
├── API 服务器启动
└── 信号处理和优雅关闭
```

### 2. API层 (api/)
```
api/
├── server.go             # HTTP API 服务器核心
├── strategy.go           #相关 API接口
├── backtest.go           #回相关 API 接口
├── debate.go             #辩相关 API 接口
├── crypto_handler.go    # 加密处理接口
├── env_handler.go        #环变量管理接口
├── guardian_routes.go   #守程序路由
├── errors.go             #错误处理
└── utils.go              #工具函数
```

### 3. 交易执行层 (trader/)
```
trader/
├── auto_trader.go        # 自动交易核心引擎
├── binance_futures.go    # Binance 期货交易器
├── bybit_trader.go       # Bybit 交易器
├── okx_trader.go         # OKX 交易器
├── bitget_trader.go      # Bitget 交易器
├── hyperliquid_trader.go # Hyperliquid 交易器
├── aster_trader.go       # Aster DEX 交易器
├── lighter_trader_v2.go  # Lighter DEX 交易器
├── interface.go          # 交易器接口定义
├── precision_manager.go  #管理器
├── error_classifier.go   #错误分类器
└── proxy_wrapper.go      # 代理包装器
```

### 4.引擎 (strategy/)
```
strategy/ (集成在 decision/ 中)
├── engine.go             #策引擎核心
├── coin_selection.go     #选择逻辑
├── data_assembly.go      # 数据组装
├── prompt_construction.go # 提示词构建
└── decision_execution.go #决执行
```

### 5.回引擎 (backtest/)
```
backtest/
├── manager.go            #回测管理器
├── runner.go             #回测运行器
├── storage.go            #回测数据存储
├── metrics.go            #性能指标计算
├── account.go            #账模拟
└── equity.go             #权计算
```

### 6.辩引擎 (debate/)
```
debate/
├── engine.go             #辩引擎核心
├── ai_roles.go           # AI角定义
├── voting_system.go      #投系统
└── consensus.go          #共算法
```

### 7. 数据存储层 (store/)
```
store/
├── store.go              #存入口点
├── gorm.go               # GORM配置
├── driver.go             # 数据库驱动
├── trader.go             # 交易器存储
├── strategy.go           #策存储
├── position.go           #持存储
├── order.go              # 订单存储
├── backtest.go           #回测数据存储
├── debate.go             #辩数据存储
├── ai_model.go           # AI模型配置存储
└── exchange.go           # 交易所配置存储
```

### 8.数据层 (market/)
```
market/
├── provider.go          # 数据提供商接口
├── coinank_api.go        # CoinAnk API
├── binance_ws.go         # Binance WebSocket (已弃用)
├── alpaca.go             # Alpaca数据
└── twelvedata.go         # TwelveData外数据
```

### 9. AI模客户端 (mcp/)
```
mcp/
├── client.go             # AI客端接口
├── deepseek.go           # DeepSeek客户端
├── openai.go             # OpenAI客户端
├── qwen.go               # 通义千问客户端
├── claude.go             # Claude客户端
├── gemini.go             # Gemini客户端
└── ollama.go             # Ollama 本地模型客户端
```

### 10.前端界面 (web/)
```
web/
├── src/
│   ├── App.tsx           #应用根组件
│   ├── main.tsx          #应用入口
│   ├── pages/            # 页面组件
│   │   ├── TraderDashboardPage.tsx # 交易仪表板
│   │   ├── StrategyStudioPage.tsx  #策工作室
│   │   ├── BacktestPage.tsx         #回测页面
│   │   ├── DebateArenaPage.tsx     #辩竞技场
│   │  └── CompetitionPage.tsx      #页面
│   ├── components/       #共组件
│   │   ├── HeaderBar.tsx           #头导航
│   │   ├── AITradersPage.tsx        # AI 交易器管理
│   │   ├── ChartTabs.tsx            # 图表组件
│   │  └── DecisionCard.tsx         #决卡片
│   ├── lib/              #工具库
│   │   ├── api.ts        # API客户端
│   │   ├── crypto.ts     # 加密工具
│   │  └── config.ts     #配置管理
│   ├── stores/           #状态管理
│   │   ├── tradersConfigStore.ts    # 交易器配置状态
│   │   └── tradersModalStore.ts    # 交易器模态框状态
│   ├── contexts/         # React 上下文
│   │   ├── AuthContext.tsx          #认证上下文
│   │   └── LanguageContext.tsx      # 语言上下文
│  └── hooks/            # 自定义钩子
│       ├── useSystemConfig.ts       #系统配置钩子
│       └── useWebSocket.ts          # WebSocket
└── package.json          #前端依赖配置
```

### 11.系统管理模块
```
admin/                    #管员系统
├── web/                  #管界面
├── controllers/          #管控制器
├── models/               #管数据理数据模型
└── routes/               #管理路由

guardian/                 #浏器览器自动化守护程序
├── main.go              #守程序入口
├── browser_automation.go #浏览器自动化
└── ai_integration.go    # AI

tools/                    #工具集
├── position_checker.go  #持检查工具
├── data_sync.go          # 数据同步工具
└── risk_monitor.go       #监控工具

tools_docker_proxy/       # Docker 代理工具
└── proxy_service.go      # 代理服务
```

##核心数据流

### 1. 实时交易流程
```
1.策配置 → 2.选择 → 3. 数据组装 → 4. AI决 → 5.控制 → 6.订单执行 → 7. 结果记录
```

### 2.回流程
```
1.策配置 → 2.历数据加载 → 3.模交易 → 4.性能计算 → 5. 结果分析 → 6.报告生成
```

### 3. AI辩流程
```
1. 问题提出 → 2.多AI角色分析 → 3.多辩论 → 4.投共识 → 5.决执行
```

##配置管理

###环境变量配置 (.env)
```
# 服务器配置
NOFX_BACKEND_PORT=8888
NOFX_FRONTEND_PORT=3300

#认证配置
JWT_SECRET=your-secret-key
DATA_ENCRYPTION_KEY=your-encryption-key
RSA_PRIVATE_KEY=your-rsa-key

# 数据库配置
DB_TYPE=sqlite
DB_PATH=data/data.db

# AI模型配置
MODEL_MAX_TOKENS_DEEPSEEK=32768
MODEL_PRICE_DEEPSEEK_INPUT=0.20

# 交易所配置
ENABLE_OI_FEATURE=true
MIN_OI_THRESHOLD_MILLIONS=15

#安全配置
TRANSPORT_ENCRYPTION=false
```

##部署架构

### 开发环境
```
┌─────────────────┐    ┌─────────────────┐
│  前端开发服    │    │   后端开发服    │
│   (Vite 3300)   │    │   (Go 8888)     │
└─────────────────┘   └─────────────────┘
         │                       │
        └───────────────────────┘
                   │
        ┌─────────▼─────────┐
         │   SQLite 数据库    │
        └───────────────────┘
```

### 生产环境 (Docker)
```
┌─────────────────────────────────────┐
│           Docker Compose            │
├─────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌───────┐ │
│  │ Nginx   │  │  Frontend│  │Backend│ │
│  │ Proxy   │  │  (React) │  │ (Go)  │ │
│  │ :80/443 │  │  :3000   │  │ :8080 │ │
│ └─────────┘  └─────────┘  └───────┘ │
│           │        │         │        │
│           └────────┴─────────┴────────┘ │
│                    │                    │
│         ┌─────────▼─────────┐          │
│         │   PostgreSQL      │          │
│         │    Database       │          │
│         └───────────────────┘          │
└─────────────────────────────────────────┘
```

##安全架构

### 数据安全
- **传输加密**: HTTPS/TLS (生产环境)
- **存储加密**: AES-256 + RSA加密
- **API密钥保护**:浏览器端 Web Crypto API 加密
- **敏感数据**: 数据库存储时加密

###认证授权
- **JWT Token**:基于 Token 的认证机制
- **权限控制**:基于角色的访问控制
- **会话管理**: Token过和刷新机制

###安全
- **CORS**:跨域资源共享控制
- **速率限制**: API 请求频率限制
- **输入验证**: 严格的参数验证
- **错误处理**:安全的错误信息返回

##监控与运维

### 日志系统
```
-应用日志: info/warn/error/debug
- 交易日志: 订单执行记录
- AI 日志:模型调用和响应记录
-系统日志:性能和健康状态
```

###监控指标
```
-系统健康: CPU/内存/磁盘使用率
- 交易性能:订单执行成功率
- AI 服务:响应时间和成功率
- 数据库: 查询性能和连接状态
```

###告警机制
```
- 交易异常:订单失败/超时告警
-系统故障: 服务不可用告警
-控制:风险阈值告警
-性能下降: 系统性能异常告警
```

##扩性设计

###扩展
- **微服务架构**:核心模块可独立部署
- **负载均衡**:支持多实例部署
- **数据库分片**:支持数据水平分片
- **缓存机制**: Redis缓存热点数据

###插化扩展
- **交易所插件**:支持新的交易所集成
- **AI模型插件**:支持新的 AI模型接入
- **策略插件**:支持自定义策略模块
- **数据源插件**:支持新的数据源接入

##性能优化

###后端优化
- **连接池**: 数据库连接池管理
- **缓存机制**:数据缓存
- **异步处理**:塞操作
- **批处理**:批数据数据处理

###前优化
- **代码分割**:按需加载组件
- **缓存策略**: HTTP缓存和本地缓存
- **虚拟滚动**:大数据列表优化
- **状态管理**:高的状态更新

## 未来发展

### 技术演进
- **云原生**: Kubernetes容编排
- **服务网格**: Istio 服务治理
- **事件驱动**: Kafka消队列
- **实时计算**: Flink流

###功能扩展
- **更多市场**: 支持更多交易品种
- **智能策略**: 机器学习策略优化
- **社交功能**:策分享和社区
- **移动应用**: iOS/Android客端

---
*文档版本: v1.0*
*最后更新: 2026年3月*