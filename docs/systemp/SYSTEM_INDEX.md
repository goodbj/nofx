# NOFX系统模块索引

##目录结构概览

```
nofx_dev/
├── main.go                    # 主程序入口
├── api/                       # HTTP API接口层
├── trader/                    # 交易执行层
├── backtest/                  #回引擎
├── debate/                    # AI辩引擎
├── decision/                  #决引擎引擎
├── strategy/                  #策模块
├── market/                    #数据服务
├── mcp/                       # AI模型客户端
├── store/                     # 数据存储层
├── config/                   #配置管理
├── auth/                     #认证授权
├── manager/                   # 管理器
├── web/                      #前端界面
├── guardian/                 #浏器览器自动化守护程序
├── tools/                    #工具集
├── admin/                    #系统
├── docs/                     # 文档目录
│   └── systemp/              #系统架构文档
└── docker/                   # Docker配置
```

##核心模块详细索引

### 1. 主程序入口 (main.go)
**文件**: `main.go`
**主要功能**:
-应用程序启动入口
-全局配置初始化
- 数据库连接管理
- 加密服务初始化
- 交易管理器启动
- API 服务器启动
- 系统信号处理

**关键组件**:
- `main()` - 主函数
- `initInstallationID()` -安装ID初始化
- `newSharedMCPClient()` -共AI客户端创建

### 2. API接口层 (api/)
**目录**: `api/`
**主要功能**: HTTP API 接口实现

####核心文件:
- `server.go` (194KB) - API 服务器核心实现
  - HTTP路配置由配置
  - 中间件管理
  - CORS支持
  - 日志中间件

- `strategy.go` (40KB) -相关 API
  - 策略创建/更新/删除
  -策配置验证
  - 策略市场接口

- `backtest.go` (23KB) -回测相关 API
  -回测任务管理
  -回测结果查询
  -回测进度监控

- `debate.go` (18KB) -辩相关 API
  -辩会话管理
  - AI角配置
  -投系统接口

- `crypto_handler.go` (2.5KB) - 加密处理
  - API密钥加密/解密
  -客端加密支持

- `env_handler.go` (16KB) -环境变量管理
  -系统配置管理
  -环境变量更新
  -配置验证

#### 主要路由组:
```
/api/v1/
├── /health                  #健检查
├── /traders/               # 交易器管理
├── /strategies/            #策管理
├── /backtest/              #回测接口
├── /debate/                #辩接口接口
├── /market/                #数据
├── /auth/                  # 认证接口
├── /config/                #配置管理
└── /system/                #系统信息
```

### 3. 交易执行层 (trader/)
**目录**: `trader/`
**主要功能**:各交易所 API和交易执行

#### 核心交易器:
- `auto_trader.go` (192KB) - 自动交易引擎核心
  - 交易周期管理
  -策执行
  -控制
  -订单管理

- `binance_futures.go` (106KB) - Binance 期货交易器
  - 期货交易接口
  - 仓位管理
  -订单执行
  -管理

- `bybit_trader.go` (40KB) - Bybit 交易器
  -现/期货交易
  -统一接口实现

- `okx_trader.go` (48KB) - OKX 交易器
  -全市场支持
  -杠交易

- `hyperliquid_trader.go` (75KB) - Hyperliquid DEX 交易器
  -合约交易
  -交易

- `aster_trader.go` (47KB) - Aster DEX 交易器
  -去化交易所
  -合约交互

- `lighter_trader_v2.go` (21KB) - Lighter DEX 交易器
  - AMM 交易
  -流性管理

####公组件:
- `interface.go` - 交易器接口定义
- `precision_manager.go` -管理
- `error_classifier.go` -错误分类处理
- `proxy_wrapper.go` - 代理包装器
- `position_sync.go` - 仓位同步

### 4.回测引擎 (backtest/)
**目录**: `backtest/`
**主要功能**:历数据回测和策略验证

####核心组件:
- `manager.go` -回测管理器
- `runner.go` -回测运行器
- `storage.go` -回测数据存储
- `metrics.go` -性能指标计算
- `account.go` -模拟
- `equity.go` -权计算
- `datafeed.go` - 数据源管理
- `ai_client.go` - AI模集成

####功能特性:
-多品种、多时间周期回测
- 实时进度监控 (SSE)
-点和恢复功能
-性能指标计算 (Sharpe、回撤、胜率等)
- AI决回放

### 5.辩引擎 (debate/)
**目录**: `debate/`
**主要功能**:多 AI模型协作决策系统

####核心组件:
- `engine.go` -辩引擎引擎核心
- AI角色系统:
  - `Bull` - 乐观交易员
  - `Bear` -交易员  
  - `Analyst` -技术分析师
  - `Contrarian` - 逆向思维者
  - `RiskManager` -控制控制专家

####工作流程:
1. 问题提出
2.多角色独立分析
3.多辩论讨论
4.权投票决策
5.共执行

### 6.决引擎 (decision/)
**目录**: `decision/`
**主要功能**: 策略执行和决策核心

####核心组件:
- `engine.go` - 策略引擎
- `coin_selection.go` - 交易对选择
- `data_assembly.go` - 数据组装
- `prompt_construction.go` - 提示词构建
- `decision_execution.go` -决执行

#### 数据流程:
```
候选币选择 →数据收集 →技术指标计算 → 
提示词构建 → AI模型调用 →决解析 → 
风险控制 →订单执行 → 结果记录
```

### 7. 数据存储层 (store/)
**目录**: `store/`
**主要功能**: 数据持久化和访问

#### 数据模型:
- `store.go` -存入口点
- `trader.go` - 交易器配置
- `strategy.go` -策配置
- `position.go` -持记录 (34KB)
- `order.go` -订单记录 (16KB)
- `backtest.go` -回测数据
- `debate.go` -辩数据数据
- `ai_model.go` - AI模型配置
- `exchange.go` - 交易所配置
- `equity.go` -权记录
- `user.go` - 用户管理

####技术实现:
- 数据库抽象层
- GORM ORM映
- SQLite/PostgreSQL支持
- 数据库迁移

### 8.市数据服务 (market/)
**目录**: `market/`
**主要功能**: 实时市场数据获取

#### 数据提供商:
- `coinank_api.go` - CoinAnk API (主要数据源)
- `binance_ws.go` - Binance WebSocket (已弃用)
- `alpaca.go` -数据 (Alpaca)
- `twelvedata.go` -外/贵金属 (TwelveData)

#### 数据类型:
- K线数据 (1分钟到1天)
- 实时价格
- 成交量数据
-持量数据
-资金费率

### 9. AI模型客户端 (mcp/)
**目录**: `mcp/`
**主要功能**:多 AI模型接入支持

####支持模型:
- `deepseek.go` - DeepSeek (成本优化首选)
- `qwen.go` - 通义千问
- `openai.go` - OpenAI (GPT系列)
- `claude.go` - Anthropic Claude
- `gemini.go` - Google Gemini
- `ollama.go` - 本地模型 (Ollama)
- `ernie.go` -百文心

####核心功能:
-统一接口抽象
- 上下文压缩优化
- 成本控制
-错误重试机制
- 使用统计

### 10.配置管理 (config/)
**目录**: `config/`
**主要功能**: 系统配置管理

####核心文件:
- `config.go` -全局配置定义
- `ports.go` -端口配置
- `guardian_config.go` -守程序配置

####配置类型:
- 服务器配置 (端口、超时)
- 数据库配置
-安全配置 (加密密钥)
- AI模型配置
- 成本显示配置
-监配置

### 11. 前端界面 (web/)
**目录**: `web/`
**主要功能**: 用户交互界面

####核心目录:
```
web/src/
├── pages/                  # 页面组件
│   ├── TraderDashboardPage.tsx (107KB)
│   ├── StrategyStudioPage.tsx (81KB)
│   ├── BacktestPage.tsx (85KB)
│   ├── DebateArenaPage.tsx (43KB)
│   └── CompetitionPage.tsx (20KB)
├── components/            #公组件
│   ├── HeaderBar.tsx (23KB)
│   ├── AITradersPage.tsx (73KB)
│   ├── ChartTabs.tsx (16KB)
│   └── DecisionCard.tsx (43KB)
├── lib/                   #工具库
│   ├── api.ts (30KB)
│   ├── crypto.ts (7KB)
│  └── config.ts
├── stores/                #状态管理
│   └── tradersModalStore.ts
└── contexts/              # React 上下文
    ├── AuthContext.tsx
   └── LanguageContext.tsx
```

####技术特点:
- React 18 + TypeScript
- Vite构建工具
- TailwindCSS样式
- Zustand状态管理
- SWR 数据获取
- Framer Motion动画

### 12.系管理统管理 (admin/)
**目录**: `admin/`
**主要功能**: 系统管理后台

####组件结构:
- `web/` -管界面界面
- `controllers/` - 业务逻辑
- `models/` - 数据模型
- `routes/` -路由配置
- `auth/` -员认证

####管功能功能:
-系统监控
- 用户管理
- 日志查看
-配置管理
-性能监控

### 13.浏器览器自动化 (guardian/)
**目录**: `guardian/`
**主要功能**:基于浏览器的 AI 自动化交易

####核心组件:
- `main.go` -守程序入口
- `browser_automation.go` -浏览器控制
- `ai_integration.go` - AI 服务集成
- `web_crawling.go` -网页爬取
- `order_execution.go` -订单执行

####应用场景:
-网页端 AI 服务交易
-手操作辅助
-特市场访问

### 14.工具集 (tools/ & tools_docker_proxy/)
**目录**: `tools/`, `tools_docker_proxy/`
**主要功能**:系统工具和辅助服务

#### 系统工具:
- `position_checker.go` - 仓位状态检查
- `data_sync.go` - 数据同步工具
- `risk_monitor.go` -监控
- `backup_tool.go` - 数据备份

#### Docker 代理服务:
- `proxy_service.go` -反代理服务
- `ssl_terminator.go` - SSL终
- `load_balancer.go` -均衡

### 15. Docker 配置 (docker/ &根 compose 文件)
**目录**: `docker/`
**主要功能**:容化部署配置

#### Dockerfiles:
- `Dockerfile.backend` -后端服务
- `Dockerfile.backend.dev` - 开发版后端
- `Dockerfile.frontend` -前端服务
- `Dockerfile.frontend.dev` - 开发版前端
- `Dockerfile.guardian` -守程序

#### Docker Compose:
- `docker-compose.dev.yml` - 开发环境
- `docker-compose.prod.yml` - 生产环境
- `docker-compose.watch.yml` -热更新环境
- `docker-compose.proxy.yml` - 代理服务
- `docker-compose.with-guardian.yml` -完整系统

### 16. 文档体系 (docs/)
**目录**: `docs/`
**主要功能**: 项目文档体系

####架文档文档:
- `systemp/SYSTEM_ARCHITECTURE.md` - 系统架构
- `systemp/SYSTEM_INDEX.md` - 系统模块索引

#### 用户文档:
- `architecture/README.md` -架总览
- `architecture/STRATEGY_MODULE.md` -策模块
- `architecture/BACKTEST_MODULE.md` -回测模块
- `architecture/DEBATE_MODULE.md` -辩模块
- `getting-started/README.md` -入指南
- `faq/README.md` -常问题

####部分文档:
- `api/` - API参文档
- `i18n/` -国际化文档
- `research/` -技术调研
- `guides/` - 用户指引

### 17.性能测试 (precision_test/)
**目录**: `precision_test/`
**主要功能**:精度测试和验证

####测试内容:
- 数值精度验证
-计准确性测试
- 交易所精度对比
- 系统性能基准

##模块依赖关系

###核心依赖图:
```
main.go
├── config/ (配置初始化)
├── store/ (数据库)
├── crypto/ (加密服务)
├── manager/ (交易管理器)
├── backtest/ (回测引擎)
├── api/ (API 服务)
│   ├── trader/
│   ├── market/
│   ├── mcp/
│  └── store/
└── guardian/ (可选)
```

###前端依赖:
```
web/src/App.tsx
├── components/
├── pages/
├── lib/api.ts
├── stores/
└── contexts/
```

## 系统启动流程

###后端启动顺序:
1.环境变量加载
2. 日志系统初始化
3.配置加载
4. 加密服务初始化
5. 数据库连接
6. 交易管理器初始化
7.回测引擎初始化
8. API 服务器启动
9.后服务启动

###前端启动流程:
1. React应用初始化
2.状态管理配置
3. API客端配置
4.路由配置
5.组件渲染
6. 数据获取
7. 用户界面显示

## 系统监控点

###检查端点:
- `/api/v1/health` -系统健康状态
- `/api/v1/system/status` - 详细系统状态
- `/api/v1/metrics` -性能指标

###监控数据:
-系统资源使用率
- API响应时间
- 数据库连接状态
- 交易执行成功率
- AI 模型调用统计

---
*文档版本: v1.0*
*最后更新: 2026年3月*