# NOFX 故障排查详细文档

## 故障排查概述

本文档提供NOFX系统常见故障的诊断方法和解决方案，涵盖从系统启动到运行维护的全生命周期故障处理。

## 系统启动故障

### 后端服务启动失败

#### 症状
```
Error: failed to connect to database
Error: invalid JWT secret
Error: port already in use
```

#### 诊断步骤

**1. 检查环境变量配置**
```bash
# 验证必需的环境变量
grep -E "(JWT_SECRET|DATA_ENCRYPTION_KEY|DB_PATH)" .env

# 检查端口占用
netstat -tuln | grep :8888
lsof -i :8888
```

**2. 数据库连接问题**
```bash
# SQLite数据库检查
ls -la data/data.db
sqlite3 data/data.db "PRAGMA integrity_check;"

# PostgreSQL连接测试
pg_isready -h localhost -p 5432
psql -h localhost -U nofx_user -d nofx -c "SELECT 1;"
```

**3. 日志分析**
```bash
# 查看详细错误日志
tail -f data/nofx.log
grep -i "error\|fatal\|panic" data/nofx.log

# Go程序崩溃信息
dmesg | grep -i "segfault\|panic"
```

#### 解决方案

**数据库文件损坏**
```bash
# 备份当前数据库
cp data/data.db data/data.db.backup

# 重建数据库
rm data/data.db
go run main.go

# 从备份恢复（如果需要）
# sqlite3 data/data.db ".restore data/data.db.backup"
```

**端口冲突解决**
```bash
# 查找占用进程
sudo netstat -tulnp | grep :8888

# 终止占用进程
sudo kill -9 $(sudo lsof -t -i:8888)

# 或修改端口配置
echo "NOFX_BACKEND_PORT=8889" >> .env
```

### 前端服务启动失败

#### 症状
```
Error: ENOENT: no such file or directory, scandir 'node_modules'
Error: Cannot find module 'react'
Error: EACCES: permission denied
```

#### 诊断步骤

**1. 依赖检查**
```bash
# 检查Node.js版本
node --version
npm --version

# 检查依赖完整性
cd web
npm ls
npm audit
```

**2. 权限问题排查**
```bash
# 检查文件权限
ls -la web/node_modules/
ls -la ~/.npm/

# 修复npm权限
sudo chown -R $(whoami) ~/.npm
sudo chown -R $(whoami) web/node_modules
```

#### 解决方案

**依赖安装问题**
```bash
# 清理并重新安装
cd web
rm -rf node_modules package-lock.json
npm cache clean --force
npm install

# 使用国内镜像源
npm config set registry https://registry.npmmirror.com
npm install
```

**端口占用处理**
```bash
# 查找前端端口占用
netstat -tuln | grep :3300
lsof -i :3300

# 临时使用其他端口
VITE_PORT=3301 npm run dev
```

## 网络连接故障

### API访问问题

#### 症状
```
Connection refused
Network timeout
CORS error
502 Bad Gateway
```

#### 诊断步骤

**1. 网络连通性测试**
```bash
# 本地连接测试
curl -v http://localhost:8888/api/v1/health
curl -v http://127.0.0.1:8888/api/v1/health

# 端口监听检查
ss -tuln | grep :8888
netstat -an | grep LISTEN | grep 8888
```

**2. 防火墙检查**
```bash
# Linux防火墙
sudo ufw status
sudo iptables -L

# Windows防火墙
netsh advfirewall show allprofiles
```

**3. DNS解析测试**
```bash
# 域名解析测试
nslookup your-domain.com
dig your-domain.com

# 测试API端点可达性
curl -v https://api.binance.com/api/v3/ping
```

#### 解决方案

**防火墙配置**
```bash
# Ubuntu防火墙开放端口
sudo ufw allow 8888/tcp
sudo ufw allow 3300/tcp
sudo ufw reload

# iptables配置
sudo iptables -A INPUT -p tcp --dport 8888 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3300 -j ACCEPT
```

**Nginx代理配置**
```nginx
# nginx配置文件检查
sudo nginx -t
sudo systemctl status nginx

# 重启Nginx
sudo systemctl restart nginx
```

### WebSocket连接失败

#### 症状
```
WebSocket connection failed
Connection closed unexpectedly
1006 connection closed abnormally
```

#### 诊断步骤

**1. WebSocket测试**
```bash
# 使用wscat测试
npm install -g wscat
wscat -c ws://localhost:8888/ws

# 浏览器控制台测试
# const ws = new WebSocket('ws://localhost:8888/ws');
# ws.onopen = () => console.log('Connected');
# ws.onerror = (err) => console.log('Error:', err);
```

**2. 代理配置检查**
```bash
# 检查代理设置
env | grep -i proxy
npm config get proxy

# 测试代理连通性
curl -x http://proxy:8080 https://api.binance.com/api/v3/ping
```

#### 解决方案

**WebSocket代理配置**
```nginx
# Nginx WebSocket支持
location /ws {
    proxy_pass http://localhost:8888;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_read_timeout 86400;
}
```

## 数据库故障

### 数据库连接失败

#### 症状
```
failed to connect to database
connection timeout
too many connections
database is locked
```

#### 诊断步骤

**1. 数据库状态检查**
```bash
# SQLite状态
sqlite3 data/data.db ".dbinfo"
sqlite3 data/data.db "PRAGMA quick_check;"

# PostgreSQL状态
pg_isready -h localhost -p 5432
psql -h localhost -U postgres -c "SELECT * FROM pg_stat_activity;"
```

**2. 连接池检查**
```bash
# 检查数据库连接数
psql -h localhost -U nofx_user -d nofx -c "SELECT count(*) FROM pg_stat_activity;"

# 检查锁等待
psql -h localhost -U nofx_user -d nofx -c "SELECT * FROM pg_locks WHERE granted = false;"
```

#### 解决方案

**连接池优化**
```go
// 数据库连接池配置优化
sqlDB.SetMaxIdleConns(10)    // 减少空闲连接
sqlDB.SetMaxOpenConns(50)    // 限制最大连接数
sqlDB.SetConnMaxLifetime(30 * time.Minute) // 缩短连接生命周期
```

**数据库维护**
```bash
# SQLite数据库维护
sqlite3 data/data.db "VACUUM;"
sqlite3 data/data.db "ANALYZE;"

# PostgreSQL数据库维护
psql -h localhost -U postgres -d nofx -c "VACUUM ANALYZE;"
psql -h localhost -U postgres -d nofx -c "REINDEX DATABASE nofx;"
```

### 数据一致性问题

#### 症状
```
data mismatch
inconsistent state
foreign key constraint failed
```

#### 诊断步骤

**1. 数据完整性检查**
```sql
-- 检查外键约束
PRAGMA foreign_key_check;

-- 检查数据一致性
SELECT t.id, COUNT(p.id) as position_count 
FROM traders t 
LEFT JOIN positions p ON t.id = p.trader_id 
GROUP BY t.id 
HAVING position_count > 0 AND t.is_running = false;
```

**2. 事务日志分析**
```bash
# 检查事务日志
grep -i "rollback\|commit\|transaction" data/nofx.log

# 数据库日志检查
tail -f /var/log/postgresql/postgresql-*.log
```

#### 解决方案

**数据修复脚本**
```sql
-- 清理不一致的交易数据
BEGIN;

-- 修复交易器状态与持仓不匹配
UPDATE traders 
SET is_running = false 
WHERE id IN (
    SELECT t.id 
    FROM traders t 
    LEFT JOIN positions p ON t.id = p.trader_id 
    WHERE t.is_running = true 
    AND p.id IS NULL
);

-- 清理孤立的订单记录
DELETE FROM orders 
WHERE trader_id NOT IN (SELECT id FROM traders);

-- 清理过期的缓存数据
DELETE FROM cache_entries WHERE expires_at < NOW();

COMMIT;
```

## AI模型集成故障

### 模型调用失败

#### 症状
```
AI model timeout
invalid API key
model not found
context length exceeded
```

#### 诊断步骤

**1. API密钥验证**
```bash
# 检查API密钥配置
grep -i "api_key" .env
echo $DEEPSEEK_API_KEY

# 测试API连接
curl -X POST https://api.deepseek.com/v1/chat/completions \
  -H "Authorization: Bearer $DEEPSEEK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"test"}],"model":"deepseek-chat"}'
```

**2. 模型可用性检查**
```bash
# 检查模型服务状态
curl -v https://api.deepseek.com/v1/models
curl -v https://api.openai.com/v1/models
```

#### 解决方案

**API密钥配置**
```bash
# 安全的API密钥设置
echo "DEEPSEEK_API_KEY=your-actual-api-key" >> .env
chmod 600 .env

# 或使用环境变量
export DEEPSEEK_API_KEY="your-api-key"
go run main.go
```

**模型回退策略**
```go
// 多模型回退机制
func callAIModelWithFallback(prompt string, models []string) (string, error) {
    for i, model := range models {
        client := getAIModelClient(model)
        response, err := client.Generate(prompt)
        if err == nil {
            return response, nil
        }
        
        log.Printf("Model %s failed: %v, trying fallback %d/%d", 
            model, err, i+1, len(models))
    }
    
    return "", errors.New("all AI models failed")
}

// 使用示例
models := []string{"deepseek", "qwen", "gpt-4"}
result, err := callAIModelWithFallback(prompt, models)
```

### 提示词相关错误

#### 症状
```
prompt too long
invalid JSON response
token limit exceeded
model response parsing failed
```

#### 解决方案

**提示词压缩优化**
```go
func optimizePromptLength(prompt string, maxTokens int) string {
    if len(prompt) <= maxTokens {
        return prompt
    }
    
    // 分析提示词结构
    sections := splitPromptSections(prompt)
    
    // 保留关键部分，压缩其他内容
    var optimizedSections []string
    totalLength := 0
    
    // 优先级：系统指令 > 关键数据 > 辅助信息
    for _, section := range priorityOrder {
        if sectionType(section) == "essential" {
            optimizedSections = append(optimizedSections, section)
            totalLength += len(section)
        } else if totalLength + len(section) <= maxTokens {
            compressed := compressSection(section, getCompressionRatio(section))
            optimizedSections = append(optimizedSections, compressed)
            totalLength += len(compressed)
        }
    }
    
    return strings.Join(optimizedSections, "\n")
}
```

## 交易所API故障

### 认证失败

#### 症状
```
invalid api key
signature mismatch
permission denied
account not found
```

#### 诊断步骤

**1. API密钥验证**
```bash
# 检查API密钥格式
echo $BINANCE_API_KEY | wc -c  # 应该是64个字符

# 测试API密钥有效性
curl -H "X-MBX-APIKEY: $BINANCE_API_KEY" \
  "https://api.binance.com/api/v3/account?timestamp=$(date +%s)000&signature=$(generate_signature)"
```

**2. 时间同步检查**
```bash
# 检查系统时间
date
ntpq -p

# 同步时间
sudo ntpdate pool.ntp.org
```

#### 解决方案

**API密钥重新配置**
```bash
# 安全地更新API密钥
# 1. 通过Web界面重新配置
# 2. 或通过API端点更新
curl -X PUT http://localhost:8888/api/v1/exchanges/binance \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"api_key": "new_key", "api_secret": "new_secret"}'
```

### 网络限制问题

#### 症状
```
403 Forbidden
429 Too Many Requests
503 Service Unavailable
connection timeout
```

#### 解决方案

**代理配置**
```bash
# 配置透明代理
echo "USE_BINANCE_PROXY=true" >> .env
echo "BINANCE_PROXY_URL=http://localhost:8081" >> .env

# 启动代理服务
cd tools_docker_proxy
docker-compose up -d
```

**请求频率控制**
```go
// 交易所API限流
type RateLimiter struct {
    requests map[string]*rate.Limiter
    weights  map[string]int // 权重配置
}

func (rl *RateLimiter) Wait(exchange, endpoint string) {
    key := fmt.Sprintf("%s:%s", exchange, endpoint)
    weight := rl.weights[endpoint]
    
    limiter := rl.getOrCreateLimiter(key)
    // 根据权重调整等待时间
    time.Sleep(time.Duration(weight) * time.Millisecond)
    limiter.Wait(context.Background())
}
```

## 系统性能问题

### 内存泄漏

#### 症状
```
memory usage keeps increasing
system becomes slow over time
out of memory errors
```

#### 诊断步骤

**1. 内存监控**
```bash
# 实时内存监控
watch -n 1 'free -h'

# 进程内存使用
ps aux --sort=-%mem | head -20

# Go程序内存分析
go tool pprof http://localhost:8888/debug/pprof/heap
```

**2. 垃圾回收分析**
```bash
# 启用GC日志
GODEBUG=gctrace=1 go run main.go

# 分析GC统计
go tool pprof -alloc_space http://localhost:8888/debug/pprof/heap
```

#### 解决方案

**内存优化**
```go
// 对象池模式减少GC压力
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processWithPool(data []byte) {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // 使用buffer处理数据
    copy(buffer, data)
    // ... 处理逻辑
}
```

### CPU使用率过高

#### 症状
```
high CPU usage
slow response times
system lag
```

#### 诊断步骤

**1. CPU性能分析**
```bash
# 实时CPU监控
top -p $(pgrep nofx)
htop

# Go CPU分析
go tool pprof http://localhost:8888/debug/pprof/profile

# 火焰图生成
go tool pprof -http=:8080 http://localhost:8888/debug/pprof/profile
```

**2. 热点代码识别**
```go
// 性能分析注释
import _ "net/http/pprof"

// 在代码中添加性能标记
func expensiveOperation() {
    defer func(start time.Time) {
        log.Printf("expensiveOperation took %v", time.Since(start))
    }(time.Now())
    
    // 执行耗时操作
}
```

#### 解决方案

**并发优化**
```go
// 限制并发数量
var semaphore = make(chan struct{}, 10) // 最大并发10个

func processConcurrently(tasks []Task) {
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {
            defer wg.Done()
            semaphore <- struct{}{}        // 获取信号量
            defer func() { <-semaphore }() // 释放信号量
            
            processTask(t)
        }(task)
    }
    
    wg.Wait()
}
```

## 日志分析和监控

### 日志级别配置

```bash
# 不同环境的日志级别
# 开发环境
LOG_LEVEL=debug

# 生产环境
LOG_LEVEL=info

# 故障排查
LOG_LEVEL=debug
```

### 关键日志模式

```bash
# 错误模式匹配
grep -i "error\|fatal\|panic" data/nofx.log

# 性能相关日志
grep -i "timeout\|slow\|performance" data/nofx.log

# 安全相关日志
grep -i "auth\|security\|unauthorized" data/nofx.log

# 交易相关日志
grep -i "trade\|order\|position" data/nofx.log
```

### 自动化监控脚本

```bash
#!/bin/bash
# system_monitor.sh

LOG_FILE="/var/log/nofx/monitor.log"
ALERT_EMAIL="admin@your-domain.com"

# 检查服务状态
check_service() {
    local service=$1
    if ! pgrep -f "$service" > /dev/null; then
        echo "$(date): $service is not running" >> $LOG_FILE
        send_alert "$service is down"
        return 1
    fi
    return 0
}

# 检查磁盘空间
check_disk_space() {
    local usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $usage -gt 80 ]; then
        echo "$(date): Disk usage is ${usage}%" >> $LOG_FILE
        send_alert "High disk usage: ${usage}%"
    fi
}

# 检查内存使用
check_memory() {
    local usage=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')
    if [ $usage -gt 85 ]; then
        echo "$(date): Memory usage is ${usage}%" >> $LOG_FILE
        send_alert "High memory usage: ${usage}%"
    fi
}

# 发送告警
send_alert() {
    local message=$1
    echo "$message" | mail -s "NOFX System Alert" $ALERT_EMAIL
}

# 主监控循环
while true; do
    check_service "nofx"
    check_disk_space
    check_memory
    sleep 60
done
```

## 常见问题快速解决

### 启动问题快速检查清单

```bash
#!/bin/bash
# quick_check.sh

echo "=== NOFX Quick Health Check ==="

# 1. 环境变量检查
echo "1. Environment variables:"
[ -f .env ] && echo "✓ .env file exists" || echo "✗ .env file missing"
grep -q "JWT_SECRET" .env && echo "✓ JWT_SECRET configured" || echo "✗ JWT_SECRET missing"

# 2. 端口检查
echo "2. Port availability:"
netstat -tuln | grep -q :8888 && echo "✓ Port 8888 available" || echo "✗ Port 8888 in use"
netstat -tuln | grep -q :3300 && echo "✓ Port 3300 available" || echo "✗ Port 3300 in use"

# 3. 数据库检查
echo "3. Database status:"
[ -f data/data.db ] && echo "✓ Database file exists" || echo "✗ Database file missing"

# 4. 依赖检查
echo "4. Dependencies:"
command -v go >/dev/null && echo "✓ Go installed" || echo "✗ Go not installed"
command -v node >/dev/null && echo "✓ Node.js installed" || echo "✗ Node.js not installed"

# 5. 服务状态
echo "5. Service status:"
pgrep -f "nofx" >/dev/null && echo "✓ Backend running" || echo "✗ Backend not running"
pgrep -f "vite" >/dev/null && echo "✓ Frontend running" || echo "✗ Frontend not running"

echo "=== Check Complete ==="
```

### 紧急恢复流程

```bash
#!/bin/bash
# emergency_recovery.sh

echo "Starting emergency recovery..."

# 1. 停止所有服务
echo "1. Stopping services..."
pkill -f nofx
pkill -f vite

# 2. 备份当前状态
echo "2. Backing up current state..."
cp -r data data.backup.$(date +%Y%m%d_%H%M%S)

# 3. 清理临时文件
echo "3. Cleaning temporary files..."
rm -rf data/tmp/*
rm -rf web/node_modules/.cache

# 4. 重启服务
echo "4. Restarting services..."
go run main.go &
cd web && npm run dev &

# 5. 验证服务状态
echo "5. Verifying service status..."
sleep 10
curl -s http://localhost:8888/api/v1/health | grep -q "healthy" && echo "✓ Backend healthy" || echo "✗ Backend unhealthy"
curl -s http://localhost:3300 | grep -q "NOFX" && echo "✓ Frontend accessible" || echo "✗ Frontend inaccessible"

echo "Emergency recovery complete!"
```

## 预防性维护

### 定期维护任务

```bash
#!/bin/bash
# maintenance.sh

# 每日维护任务
daily_maintenance() {
    echo "Running daily maintenance..."
    
    # 清理日志文件
    find data/ -name "*.log" -mtime +7 -delete
    
    # 数据库维护
    sqlite3 data/data.db "VACUUM;"
    sqlite3 data/data.db "ANALYZE;"
    
    # 清理缓存
    rm -rf data/cache/*
    
    # 更新依赖
    go mod tidy
    cd web && npm audit fix
}

# 每周维护任务
weekly_maintenance() {
    echo "Running weekly maintenance..."
    
    # 系统更新
    sudo apt update && sudo apt upgrade -y
    
    # 备份数据库
    cp data/data.db backups/data_$(date +%Y%m%d).db
    
    # 性能分析
    go tool pprof http://localhost:8888/debug/pprof/heap > heap_profile.txt
}

# 每月维护任务
monthly_maintenance() {
    echo "Running monthly maintenance..."
    
    # 完整系统备份
    tar -czf backups/full_backup_$(date +%Y%m%d).tar.gz data/ .env web/
    
    # 安全扫描
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec ./...
    
    # 依赖安全检查
    npm audit
}

# 根据传入参数执行相应维护任务
case $1 in
    daily)   daily_maintenance ;;
    weekly)  weekly_maintenance ;;
    monthly) monthly_maintenance ;;
    *)       echo "Usage: $0 {daily|weekly|monthly}" ;;
esac
```

---
*故障排查文档版本: v1.0*
*最后更新: 2026年3月*