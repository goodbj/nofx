# NOFX系统快速参考指南

## 系统启动命令

### 开发环境启动

####标准开发版 (无头模式)
```bash
# Windows
start_nofx_dev_docker.bat

#访问地址
#前端: http://localhost:3300
#后端: http://localhost:8888
```

####显示分组服务 (带浏览器显示)
```bash
# Windows -快启动
quick_start_display.bat

# 或使用完整启动脚本
start_nofx_dev_display.bat

# 或使用手动启动脚本（推荐，解决端口映射问题）
start_nofx_dev_display_manual.bat
```

#### 透明代理服务
```bash
# 启动代理服务
deploy_proxy_service.bat

#访问地址
# 代理: http://localhost:8081
#检查: http://localhost:8081/health
```

### 生产环境部署

#### 一键部署
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/NoFxAiOS/nofx/main/install.sh | bash

#访问地址: http://127.0.0.1:3300
```

#### Docker Compose部署
```bash
# 下载并启动
curl -O https://raw.githubusercontent.com/NoFxAiOS/nofx/main/docker-compose.prod.yml
docker compose -f docker-compose.prod.yml up -d

#管命令
docker compose -f docker-compose.prod.yml logs -f    # 查看日志
docker compose -f docker-compose.prod.yml restart     # 重启服务
docker compose -f docker-compose.prod.yml down       #停服务
```

## 环境变量配置

###核心配置 (.env)
```bash
# 服务器端口
NOFX_BACKEND_PORT=8888
NOFX_FRONTEND_PORT=3300

#认证配置 (必须)
JWT_SECRET=your-jwt-secret-change-this-in-production
DATA_ENCRYPTION_KEY=your-base64-encoded-32-byte-key
RSA_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nYOUR_KEY_HERE\n-----END RSA PRIVATE KEY-----

# 数据库配置
DB_TYPE=sqlite
DB_PATH=data/data.db

# AI模型配置
MODEL_MAX_TOKENS_DEEPSEEK=32768
MODEL_PRICE_DEEPSEEK_INPUT=0.20
MODEL_PRICE_DEEPSEEK_OUTPUT=0.80

#安全配置
TRANSPORT_ENCRYPTION=false
```

### 生成密钥命令
```bash
# 生成 JWT密钥
openssl rand -base64 32

# 生成 AES 加密密钥
openssl rand -base64 32

# 生成 RSA密对
openssl genrsa 2048
```

## API接口快速参考

###认证接口
```
POST /api/v1/auth/login
POST /api/v1/auth/register
POST /api/v1/auth/refresh
```

### 交易器管理
```
GET    /api/v1/traders
POST   /api/v1/traders
GET    /api/v1/traders/{id}
PUT    /api/v1/traders/{id}
DELETE /api/v1/traders/{id}
POST   /api/v1/traders/{id}/start
POST   /api/v1/traders/{id}/stop
```

###策管理
```
GET    /api/v1/strategies
POST   /api/v1/strategies
GET    /api/v1/strategies/{id}
PUT    /api/v1/strategies/{id}
DELETE /api/v1/strategies/{id}
```

###回接口
```
POST   /api/v1/backtest/run
GET    /api/v1/backtest/runs
GET    /api/v1/backtest/runs/{id}
DELETE /api/v1/backtest/runs/{id}
```

### 数据监控
```
GET /api/v1/health
GET /api/v1/system/status
GET /api/v1/system/metrics
GET /api/v1/config/public
```

##目结构速查

### 核心模块目录
```
├── main.go          # 主程序入口
├── api/             # API接口层
├── trader/          # 交易执行层
├── backtest/        #回测引擎
├── debate/          #辩引擎
├── decision/        #决引擎引擎
├── market/          # 数据服务
├── mcp/             # AI模型客户端
├── store/           # 数据存储
├── web/             #前端界面
├── config/          #配置管理
├── guardian/        #浏器览器自动化
└── docs/systemp/    #系统文档
```

###配置文件位置
```
├── .env             # 环境变量配置
├── .env.example     #配置模板
├── docker-compose.*.yml  # Docker配置
└── web/.env         #前端环境配置
```

## 数据库表结构

###核心数据表
```
traders          # 交易器配置
strategies       #策配置
positions        #持记录
orders           # 订单记录
backtest_runs    #回测运行记录
debate_sessions  #辩会话记录
ai_models        # AI模型配置
exchanges        # 交易所配置
```

### 数据查询示例
```sql
-- 查询活跃交易器
SELECT * FROM traders WHERE is_running = true;

-- 查询最新持仓
SELECT * FROM positions ORDER BY created_at DESC LIMIT 10;

-- 查询回测结果
SELECT * FROM backtest_runs WHERE status = 'completed' ORDER BY created_at DESC;
```

##常问题排查

###启动问题
```bash
#检查端口占用
netstat -an | grep :8888
netstat -an | grep :3300

# 检查Docker服务
docker info
docker compose version

# 查看日志
docker logs nofx-dev
tail -f data/nofx.log
```

### 数据库问题
```bash
#检查数据库文件
ls -la data/data.db

# 重建数据库
rm data/data.db
go run main.go

# 数据库迁移
go run scripts/migrate.go
```

###连接问题
```bash
#测试API连接
curl http://localhost:8888/api/v1/health

#测试前端连接
curl http://localhost:3300

# 检查防火墙
netstat -an | grep LISTEN
```

##性能监控

### 系统资源监控
```bash
# CPU和内存使用
top
htop

#使用情况
df -h
du -sh data/

#连接状态
netstat -an | grep :8888
ss -tuln | grep :8888
```

###应用性能监控
```bash
# API响应时间测试
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8888/api/v1/health

# 数据库查询性能
sqlite3 data/data.db "EXPLAIN QUERY PLAN SELECT * FROM positions LIMIT 10;"
```

## 日志查看

###后端日志
```bash
# 实时查看日志
tail -f data/nofx.log

#按过滤
grep "ERROR" data/nofx.log
grep "WARN" data/nofx.log

# 按模块过滤
grep "trader" data/nofx.log
grep "api" data/nofx.log
```

###前端日志
```bash
# 开发者工具控制台
# F12 → Console标签页

# 查看网络请求
# F12 → Network标签页
```

##安全配置检查

###必须配置的安全项
```bash
#检查JWT密钥
grep "JWT_SECRET" .env

#检查加密密钥
grep "DATA_ENCRYPTION_KEY" .env

#检查RSA密钥
grep "RSA_PRIVATE_KEY" .env
```

### 生产环境安全建议
```
✓ 使用强密码和密钥
✓启HTTPS
✓配置防火墙规则
✓定备份数据
✓监控异常访问
✓ 更新依赖包
```

##系统维护

###定期维护任务
```bash
#清理日志文件
find data/ -name "*.log" -mtime +30 -delete

# 数据库优化
sqlite3 data/data.db "VACUUM;"

# 依赖包更新
go mod tidy
cd web && npm update

#系统备份
tar -czf backup_$(date +%Y%m%d).tar.gz data/ .env
```

###系统升级
```bash
#最新代码
git pull origin main

# 更新依赖
go mod download
cd web && npm install

# 重启服务
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

## 开发工具推荐

###必开发工具
```
IDE: VS Code 或 GoLand
数据库工具: SQLite Browser 或 DBeaver
API测试: Postman 或 Insomnia
版本控制: Git
容器工具: Docker Desktop
```

### VS Code推插件
```
Go
Go Test Explorer
Docker
ESLint
Prettier
Tailwind CSS IntelliSense
```

##紧故障处理

### 系统无响应
```bash
#强重启服务
docker compose -f docker-compose.dev.yml down
docker compose -f docker-compose.dev.yml up -d

#检查系统资源
free -h
df -h
```

### 数据库损坏
```bash
#备份当前数据
cp data/data.db data/data.db.backup

# 重建数据库
rm data/data.db
go run main.go
```

###服务不可用
```bash
# 检查网络配置
ipconfig /all  # Windows
ifconfig       # Linux/macOS

# 检查DNS解析
nslookup api.binance.com
```

## 系统文档位置

###技术文档
```
docs/systemp/SYSTEM_ARCHITECTURE.md  # 系统架构
docs/systemp/SYSTEM_INDEX.md         #模块索引
docs/architecture/README.md          #架文档文档
docs/getting-started/README.md        #入指南
```

### 用户文档
```
README.md                            # 项目主文档
docs/faq/README.md                   #常问题
docs/guides/                         # 使用指南
docs/i18n/                           #多语言文档
```

---
*文档版本: v1.0*
*最后更新: 2026年3月*
*适用于 NOFX v1.0*