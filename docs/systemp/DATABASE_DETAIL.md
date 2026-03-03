# NOFX 数据库设计详细文档

## 数据库概述

NOFX采用关系型数据库设计，支持SQLite（开发环境）和PostgreSQL（生产环境）两种数据库引擎。数据库设计遵循第三范式，确保数据一致性和完整性。

## 数据库架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        用户管理模块                          │
├─────────────────────────────────────────────────────────────┤
│ users (用户表)                                               │
└─────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────┐
│                        核心配置模块                          │
├─────────────────────────────────────────────────────────────┤
│ ai_models (AI模型配置)                                       │
│ exchanges (交易所配置)                                       │
│ strategies (策略配置)                                        │
│ traders (交易器配置)                                         │
└─────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────┐
│                        交易数据模块                          │
├─────────────────────────────────────────────────────────────┤
│ positions (持仓记录)                                         │
│ orders (订单记录)                                            │
│ trades (交易记录)                                            │
└─────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────┐
│                        分析引擎模块                          │
├─────────────────────────────────────────────────────────────┤
│ backtest_runs (回测运行记录)                                 │
│ backtest_results (回测结果)                                  │
│ debate_sessions (辩论会话)                                   │
│ debate_rounds (辩论轮次)                                     │
└─────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────┐
│                        系统管理模块                          │
├─────────────────────────────────────────────────────────────┤
│ system_configs (系统配置)                                    │
│ system_logs (系统日志)                                       │
│ equity_records (权益记录)                                    │
└─────────────────────────────────────────────────────────────┘
```

## 核心数据表详细设计

### 1. 用户管理表

#### users (用户表)
```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明**:
- `id`: 用户唯一标识符 (UUID)
- `username`: 用户名，唯一约束
- `email`: 邮箱地址，唯一约束
- `password_hash`: 密码哈希值
- `is_active`: 账户是否激活
- `last_login`: 最后登录时间
- `created_at`: 创建时间
- `updated_at`: 更新时间

### 2. AI模型配置表

#### ai_models (AI模型配置)
```sql
CREATE TABLE ai_models (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    api_key TEXT ENCRYPTED,
    custom_api_url TEXT,
    custom_model_name TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明**:
- `provider`: AI提供商 (deepseek, openai, qwen, claude等)
- `api_key`: API密钥 (加密存储)
- `custom_api_url`: 自定义API地址
- `config`: JSON格式的额外配置
- `is_default`: 是否为默认模型

### 3. 交易所配置表

#### exchanges (交易所配置)
```sql
CREATE TABLE exchanges (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name VARCHAR(50) NOT NULL,
    exchange_type VARCHAR(20) NOT NULL, -- cex, dex
    api_key TEXT ENCRYPTED,
    api_secret TEXT ENCRYPTED,
    api_passphrase TEXT ENCRYPTED,
    custom_endpoint TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**字段说明**:
- `exchange_type`: 交易所类型 (CEX:中心化, DEX:去中心化)
- `api_key/api_secret`: API认证信息 (加密存储)
- `custom_endpoint`: 自定义API端点

### 4. 策略配置表

#### strategies (策略配置)
```sql
CREATE TABLE strategies (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    coin_source JSONB NOT NULL,
    indicators JSONB NOT NULL,
    risk_control JSONB NOT NULL,
    ai_prompt JSONB NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**JSON字段结构**:

**coin_source**:
```json
{
  "type": "static|ai500|oi_ranking|mixed",
  "coins": ["BTCUSDT", "ETHUSDT"],
  "config": {
    "min_oi_threshold": 15,
    "max_coins": 10
  }
}
```

**indicators**:
```json
{
  "ema_periods": [9, 21, 55],
  "enable_macd": true,
  "enable_rsi": true,
  "enable_atr": true,
  "enable_volume": true,
  "enable_oi": false,
  "enable_funding_rate": false
}
```

**risk_control**:
```json
{
  "max_position_size": 0.1,
  "stop_loss_percent": 5.0,
  "take_profit_percent": 10.0,
  "max_leverage": 10,
  "max_daily_trades": 20
}
```

### 5. 交易器配置表

#### traders (交易器配置)
```sql
CREATE TABLE traders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    exchange_id TEXT NOT NULL REFERENCES exchanges(id),
    ai_model_id TEXT NOT NULL REFERENCES ai_models(id),
    strategy_id TEXT NOT NULL REFERENCES strategies(id),
    config JSONB NOT NULL,
    is_running BOOLEAN DEFAULT FALSE,
    last_started TIMESTAMP,
    last_stopped TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**config字段示例**:
```json
{
  "trading_pair": "BTCUSDT",
  "leverage": 10,
  "position_size": 0.01,
  "interval_minutes": 5,
  "max_positions": 3
}
```

### 6. 持仓记录表

#### positions (持仓记录)
```sql
CREATE TABLE positions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    trader_id TEXT NOT NULL REFERENCES traders(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- long, short
    quantity DECIMAL(20,8) NOT NULL,
    entry_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8),
    leverage INTEGER DEFAULT 1,
    margin DECIMAL(20,8),
    unrealized_pnl DECIMAL(20,8) DEFAULT 0,
    realized_pnl DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'open', -- open, closed, liquidated
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP,
    close_reason VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 7. 订单记录表

#### orders (订单记录)
```sql
CREATE TABLE orders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    trader_id TEXT NOT NULL REFERENCES traders(id),
    position_id TEXT REFERENCES positions(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- buy, sell
    order_type VARCHAR(20) NOT NULL, -- market, limit, stop_loss
    quantity DECIMAL(20,8) NOT NULL,
    price DECIMAL(20,8),
    stop_price DECIMAL(20,8),
    status VARCHAR(20) NOT NULL, -- pending, filled, cancelled, rejected
    exchange_order_id TEXT,
    filled_quantity DECIMAL(20,8) DEFAULT 0,
    filled_price DECIMAL(20,8),
    commission DECIMAL(20,8) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    filled_at TIMESTAMP
);
```

### 8. 回测数据表

#### backtest_runs (回测运行记录)
```sql
CREATE TABLE backtest_runs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    strategy_id TEXT NOT NULL REFERENCES strategies(id),
    name VARCHAR(100),
    symbols TEXT[] NOT NULL,
    timeframe VARCHAR(10) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    initial_capital DECIMAL(20,2) NOT NULL,
    config JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, running, completed, failed
    progress INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### backtest_results (回测结果)
```sql
CREATE TABLE backtest_results (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES backtest_runs(id),
    symbol VARCHAR(20) NOT NULL,
    total_return DECIMAL(10,4),
    annual_return DECIMAL(10,4),
    sharpe_ratio DECIMAL(10,4),
    max_drawdown DECIMAL(10,4),
    win_rate DECIMAL(5,2),
    total_trades INTEGER,
    profitable_trades INTEGER,
    avg_win DECIMAL(10,4),
    avg_loss DECIMAL(10,4),
    profit_factor DECIMAL(10,4),
    equity_curve JSONB,
    trade_log JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 9. 辩论数据表

#### debate_sessions (辩论会话)
```sql
CREATE TABLE debate_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    topic TEXT NOT NULL,
    symbols TEXT[] NOT NULL,
    config JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, running, completed, failed
    current_round INTEGER DEFAULT 0,
    max_rounds INTEGER NOT NULL,
    consensus JSONB,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### debate_rounds (辩论轮次)
```sql
CREATE TABLE debate_rounds (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES debate_sessions(id),
    round_number INTEGER NOT NULL,
    model_name VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL, -- bull, bear, analyst, contrarian, risk_manager
    response TEXT NOT NULL,
    vote JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 10. 系统管理表

#### system_configs (系统配置)
```sql
CREATE TABLE system_configs (
    id TEXT PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### system_logs (系统日志)
```sql
CREATE TABLE system_logs (
    id TEXT PRIMARY KEY,
    level VARCHAR(10) NOT NULL, -- debug, info, warn, error
    module VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    details JSONB,
    user_id TEXT REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### equity_records (权益记录)
```sql
CREATE TABLE equity_records (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    trader_id TEXT REFERENCES traders(id),
    timestamp TIMESTAMP NOT NULL,
    equity DECIMAL(20,2) NOT NULL,
    balance DECIMAL(20,2) NOT NULL,
    unrealized_pnl DECIMAL(20,2) DEFAULT 0,
    realized_pnl DECIMAL(20,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 索引设计

### 核心索引
```sql
-- 用户相关索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 交易器相关索引
CREATE INDEX idx_traders_user_id ON traders(user_id);
CREATE INDEX idx_traders_status ON traders(is_running);
CREATE INDEX idx_traders_created_at ON traders(created_at);

-- 持仓相关索引
CREATE INDEX idx_positions_trader_id ON positions(trader_id);
CREATE INDEX idx_positions_symbol ON positions(symbol);
CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_opened_at ON positions(opened_at);

-- 订单相关索引
CREATE INDEX idx_orders_trader_id ON orders(trader_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

-- 回测相关索引
CREATE INDEX idx_backtest_runs_user_id ON backtest_runs(user_id);
CREATE INDEX idx_backtest_runs_status ON backtest_runs(status);
CREATE INDEX idx_backtest_runs_created_at ON backtest_runs(created_at);

-- 日志相关索引
CREATE INDEX idx_system_logs_level ON system_logs(level);
CREATE INDEX idx_system_logs_module ON system_logs(module);
CREATE INDEX idx_system_logs_created_at ON system_logs(created_at);
```

## 数据库迁移

### 版本控制
使用GORM的自动迁移功能：

```go
// 自动迁移所有表
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &User{},
        &AIModel{},
        &Exchange{},
        &Strategy{},
        &Trader{},
        &Position{},
        &Order{},
        &BacktestRun{},
        &BacktestResult{},
        &DebateSession{},
        &DebateRound{},
        &SystemConfig{},
        &SystemLog{},
        &EquityRecord{},
    )
}
```

### 手动迁移脚本
```sql
-- 数据库初始化脚本
-- version: 1.0.0

BEGIN;

-- 创建表结构
-- (此处包含上述所有CREATE TABLE语句)

-- 插入初始数据
INSERT INTO system_configs (id, key, value, description) VALUES
('config_001', 'version', '1.0.0', '系统版本号'),
('config_002', 'max_users', '10', '最大用户数限制'),
('config_003', 'registration_enabled', 'true', '是否启用注册');

-- 创建默认用户
INSERT INTO users (id, username, email, password_hash) VALUES
('user_admin', 'admin', 'admin@nofx.local', 'hashed_password');

COMMIT;
```

## 数据备份策略

### 备份频率
- **生产环境**: 每日全量备份 + 每小时增量备份
- **开发环境**: 每周备份

### 备份脚本示例
```bash
#!/bin/bash
# 数据库备份脚本

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/nofx"
DB_NAME="nofx"

# SQLite备份
if [ "$DB_TYPE" = "sqlite" ]; then
    cp data/data.db ${BACKUP_DIR}/data_${DATE}.db
    zip -j ${BACKUP_DIR}/backup_${DATE}.zip ${BACKUP_DIR}/data_${DATE}.db
    rm ${BACKUP_DIR}/data_${DATE}.db
fi

# PostgreSQL备份
if [ "$DB_TYPE" = "postgres" ]; then
    pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > ${BACKUP_DIR}/dump_${DATE}.sql
    gzip ${BACKUP_DIR}/dump_${DATE}.sql
fi

# 清理旧备份（保留30天）
find ${BACKUP_DIR} -name "backup_*.zip" -mtime +30 -delete
find ${BACKUP_DIR} -name "dump_*.sql.gz" -mtime +30 -delete
```

## 性能优化

### 查询优化
```sql
-- 分析查询性能
EXPLAIN QUERY PLAN SELECT * FROM positions WHERE trader_id = ? AND status = 'open';

-- 创建复合索引优化常用查询
CREATE INDEX idx_positions_trader_status ON positions(trader_id, status);
CREATE INDEX idx_orders_trader_status ON orders(trader_id, status);
CREATE INDEX idx_backtest_user_status ON backtest_runs(user_id, status);
```

### 数据库配置优化

**SQLite配置**:
```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = MEMORY;
```

**PostgreSQL配置**:
```sql
-- 连接池配置
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
```

## 数据安全

### 加密存储
- 敏感字段使用AES-256加密
- API密钥在数据库中加密存储
- 用户密码使用bcrypt哈希

### 访问控制
- 基于用户ID的数据隔离
- 行级安全策略
- 定期权限审计

## 监控指标

### 数据库健康检查
```sql
-- 检查表大小
SELECT 
    name AS table_name,
    pg_size_pretty(pg_total_relation_size(C.oid)) AS total_size
FROM pg_class C
LEFT JOIN pg_namespace N ON (N.oid = C.relnamespace)
WHERE nspname NOT IN ('pg_catalog', 'information_schema')
AND C.relkind = 'r'
ORDER BY pg_total_relation_size(C.oid) DESC;

-- 检查连接数
SELECT count(*) FROM pg_stat_activity;

-- 检查慢查询
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

---
*数据库文档版本: v1.0*
*最后更新: 2026年3月*