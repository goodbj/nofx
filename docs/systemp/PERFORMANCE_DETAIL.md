# NOFX 性能优化详细文档

## 性能优化概述

NOFX系统通过多维度性能优化策略，包括数据库优化、缓存机制、并发处理、资源管理等，确保系统在高负载下仍能保持稳定高效的运行表现。

## 性能监控指标

### 核心性能指标

#### 系统级指标
```
CPU使用率: < 80%
内存使用率: < 85%
磁盘I/O: < 1000 IOPS
网络带宽: < 80% 使用率
```

#### 应用级指标
```
API响应时间: < 200ms (95th percentile)
数据库查询时间: < 50ms (平均)
WebSocket连接数: < 10000
并发用户数: < 1000
```

#### 业务指标
```
交易执行延迟: < 100ms
AI模型响应时间: < 30s
回测计算速度: > 1000 candles/秒
数据同步延迟: < 1s
```

### 性能监控工具配置

```yaml
# Prometheus监控配置
prometheus.yml:
  global:
    scrape_interval: 15s
    evaluation_interval: 15s
  
  scrape_configs:
    - job_name: 'nofx-backend'
      static_configs:
        - targets: ['localhost:8888']
      metrics_path: '/api/v1/metrics'
      
    - job_name: 'nofx-frontend'
      static_configs:
        - targets: ['localhost:3300']
      
    - job_name: 'postgresql'
      static_configs:
        - targets: ['localhost:9187']
      
    - job_name: 'node-exporter'
      static_configs:
        - targets: ['localhost:9100']
```

## 数据库性能优化

### 查询优化

#### 索引优化策略
```sql
-- 核心查询索引
CREATE INDEX CONCURRENTLY idx_positions_trader_status 
ON positions(trader_id, status) 
WHERE status = 'open';

CREATE INDEX CONCURRENTLY idx_orders_trader_created 
ON orders(trader_id, created_at DESC);

CREATE INDEX CONCURRENTLY idx_backtest_user_status 
ON backtest_runs(user_id, status, created_at DESC);

-- 复合索引优化
CREATE INDEX CONCURRENTLY idx_equity_user_timestamp 
ON equity_records(user_id, timestamp DESC);

-- 部分索引减少存储开销
CREATE INDEX CONCURRENTLY idx_active_traders 
ON traders(id) 
WHERE is_running = true;
```

#### 查询优化示例
```go
// 优化前：N+1查询问题
func GetTraderWithPositions(traderID string) (*Trader, error) {
    var trader Trader
    if err := db.First(&trader, "id = ?", traderID).Error; err != nil {
        return nil, err
    }
    
    // N+1查询问题
    var positions []Position
    for _, position := range trader.Positions {
        db.Where("id = ?", position.ID).First(&position)
    }
    
    return &trader, nil
}

// 优化后：预加载关联数据
func GetTraderWithPositionsOptimized(traderID string) (*Trader, error) {
    var trader Trader
    err := db.Preload("Positions", func(db *gorm.DB) *gorm.DB {
        return db.Where("status = ?", "open").Order("opened_at DESC")
    }).Preload("Orders").
       First(&trader, "id = ?", traderID).Error
    
    return &trader, err
}
```

### 连接池优化

```go
// 数据库连接池配置
func NewOptimizedDB(config DBConfig) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
        PrepareStmt: true, // 预编译SQL语句
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池优化配置
    sqlDB.SetMaxIdleConns(20)           // 最大空闲连接数
    sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
    
    return db, nil
}
```

### 数据库分片策略

```go
// 水平分片实现
type ShardingStrategy interface {
    GetShardKey(userID string) string
    GetTableName(baseTable string, shardKey string) string
}

type UserIDSharding struct{}

func (s *UserIDSharding) GetShardKey(userID string) string {
    // 基于用户ID的哈希分片
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    return fmt.Sprintf("shard_%d", hash.Sum32()%10)
}

func (s *UserIDSharding) GetTableName(baseTable, shardKey string) string {
    return fmt.Sprintf("%s_%s", baseTable, shardKey)
}
```

## 缓存优化策略

### 多级缓存架构

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层缓存                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   本地缓存   │  │   共享缓存   │  │   分布式    │         │
│  │  sync.Map   │  │   Redis     │  │   缓存集群   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                    数据层缓存                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  数据库     │  │  查询结果   │  │  索引缓存   │         │
│  │  BufferPool │  │  缓存      │  │   优化      │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Redis缓存实现

```go
// Redis缓存配置
type CacheConfig struct {
    Addr         string        `json:"addr"`
    Password     string        `json:"password"`
    DB           int           `json:"db"`
    PoolSize     int           `json:"pool_size"`
    MinIdleConns int           `json:"min_idle_conns"`
    MaxRetries   int           `json:"max_retries"`
    Timeout      time.Duration `json:"timeout"`
}

// 缓存服务实现
type RedisCache struct {
    client *redis.Client
    config CacheConfig
}

func NewRedisCache(config CacheConfig) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:         config.Addr,
        Password:     config.Password,
        DB:           config.DB,
        PoolSize:     config.PoolSize,
        MinIdleConns: config.MinIdleConns,
        MaxRetries:   config.MaxRetries,
        DialTimeout:  config.Timeout,
        ReadTimeout:  config.Timeout,
        WriteTimeout: config.Timeout,
    })
    
    return &RedisCache{
        client: client,
        config: config,
    }
}

// 智能缓存策略
func (c *RedisCache) GetWithFallback(key string, fallback func() (interface{}, error), ttl time.Duration) (interface{}, error) {
    // 尝试从缓存获取
    val, err := c.client.Get(context.Background(), key).Result()
    if err == nil {
        var result interface{}
        if err := json.Unmarshal([]byte(val), &result); err == nil {
            return result, nil
        }
    }
    
    // 缓存未命中，执行回退函数
    result, err := fallback()
    if err != nil {
        return nil, err
    }
    
    // 异步更新缓存
    go func() {
        if data, err := json.Marshal(result); err == nil {
            c.client.Set(context.Background(), key, data, ttl)
        }
    }()
    
    return result, nil
}
```

### 本地缓存优化

```go
// LRU本地缓存
type LRUCache struct {
    cache    *lru.Cache
    mutex    sync.RWMutex
    maxSize  int
    ttl      time.Duration
}

func NewLRUCache(maxSize int, ttl time.Duration) *LRUCache {
    cache, _ := lru.New(maxSize)
    return &LRUCache{
        cache:   cache,
        maxSize: maxSize,
        ttl:     ttl,
    }
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if val, ok := c.cache.Get(key); ok {
        if item, ok := val.(*cacheItem); ok {
            if time.Now().Before(item.expireAt) {
                return item.data, true
            }
            // 过期，删除
            c.cache.Remove(key)
        }
    }
    return nil, false
}

func (c *LRUCache) Set(key string, value interface{}) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    item := &cacheItem{
        data:     value,
        expireAt: time.Now().Add(c.ttl),
    }
    c.cache.Add(key, item)
}
```

## 并发性能优化

### Goroutine池管理

```go
// Goroutine池实现
type WorkerPool struct {
    workers    int
    maxWorkers int
    queue      chan Job
    semaphore  chan struct{}
    wg         sync.WaitGroup
}

type Job func() error

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        maxWorkers: workers * 2,
        queue:      make(chan Job, queueSize),
        semaphore:  make(chan struct{}, workers),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    for job := range wp.queue {
        wp.semaphore <- struct{}{} // 获取信号量
        if err := job(); err != nil {
            log.Printf("Job failed: %v", err)
        }
        <-wp.semaphore // 释放信号量
    }
}

func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.queue <- job:
        return nil
    default:
        // 队列满时动态扩展
        if len(wp.queue) >= cap(wp.queue) && wp.workers < wp.maxWorkers {
            wp.scaleUp()
        }
        return errors.New("job queue is full")
    }
}
```

### 批处理优化

```go
// 批处理引擎
type BatchProcessor struct {
    batchSize    int
    flushTimeout time.Duration
    buffer       []interface{}
    mutex        sync.Mutex
    processor    func([]interface{}) error
}

func NewBatchProcessor(batchSize int, flushTimeout time.Duration, processor func([]interface{}) error) *BatchProcessor {
    bp := &BatchProcessor{
        batchSize:    batchSize,
        flushTimeout: flushTimeout,
        processor:    processor,
        buffer:       make([]interface{}, 0, batchSize),
    }
    
    // 启动定时刷新
    go bp.periodicFlush()
    
    return bp
}

func (bp *BatchProcessor) Add(item interface{}) {
    bp.mutex.Lock()
    defer bp.mutex.Unlock()
    
    bp.buffer = append(bp.buffer, item)
    
    // 达到批次大小时立即处理
    if len(bp.buffer) >= bp.batchSize {
        bp.flush()
    }
}

func (bp *BatchProcessor) flush() {
    if len(bp.buffer) == 0 {
        return
    }
    
    // 复制缓冲区数据
    items := make([]interface{}, len(bp.buffer))
    copy(items, bp.buffer)
    bp.buffer = bp.buffer[:0]
    
    // 异步处理
    go func() {
        if err := bp.processor(items); err != nil {
            log.Printf("Batch processing failed: %v", err)
        }
    }()
}
```

## 内存优化

### 对象池模式

```go
// 对象池实现
type ObjectPool struct {
    pool sync.Pool
}

func NewObjectPool() *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &ReusableObject{
                    buffer: make([]byte, 1024),
                    data:   make(map[string]interface{}),
                }
            },
        },
    }
}

type ReusableObject struct {
    buffer []byte
    data   map[string]interface{}
    reset  func()
}

func (op *ObjectPool) Get() *ReusableObject {
    obj := op.pool.Get().(*ReusableObject)
    if obj.reset != nil {
        obj.reset()
    }
    return obj
}

func (op *ObjectPool) Put(obj *ReusableObject) {
    op.pool.Put(obj)
}
```

### 内存分析和优化

```go
// 内存使用监控
func monitorMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
    log.Printf("TotalAlloc = %v MiB", bToMb(m.TotalAlloc))
    log.Printf("Sys = %v MiB", bToMb(m.Sys))
    log.Printf("NumGC = %v", m.NumGC)
    
    // 内存使用过高时触发GC
    if m.Alloc > 100*1024*1024 { // 100MB
        runtime.GC()
    }
}

func bToMb(b uint64) uint64 {
    return b / 1024 / 1024
}
```

## 网络性能优化

### HTTP连接优化

```go
// 高性能HTTP客户端
func NewOptimizedHTTPClient() *http.Client {
    transport := &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
        MaxConnsPerHost:       50,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        // 启用HTTP/2
        ForceAttemptHTTP2: true,
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}
```

### WebSocket优化

```go
// WebSocket连接池
type WebSocketPool struct {
    connections map[string]*websocket.Conn
    mutex       sync.RWMutex
    maxConns    int
}

func (wp *WebSocketPool) GetConnection(userID string) (*websocket.Conn, error) {
    wp.mutex.RLock()
    conn, exists := wp.connections[userID]
    wp.mutex.RUnlock()
    
    if exists {
        // 检查连接是否有效
        if err := conn.WriteMessage(websocket.PingMessage, nil); err == nil {
            return conn, nil
        }
        // 连接失效，移除
        wp.RemoveConnection(userID)
    }
    
    // 创建新连接
    return wp.CreateConnection(userID)
}
```

## AI模型性能优化

### 模型调用优化

```go
// AI模型调用池
type AIModelPool struct {
    models   map[string]*ModelInstance
    mutex    sync.RWMutex
    maxConns int
}

type ModelInstance struct {
    client     *http.Client
    baseURL    string
    apiKey     string
    semaphore  chan struct{}
    lastUsed   time.Time
}

// 上下文压缩优化
func compressContext(prompt string, maxTokens int) string {
    if len(prompt) <= maxTokens {
        return prompt
    }
    
    // 保留关键部分：系统提示 + 最新数据
    lines := strings.Split(prompt, "\n")
    if len(lines) <= 10 {
        return prompt[:maxTokens]
    }
    
    // 保留前20%和后80%的内容
    keepStart := len(lines) / 5
    compressed := make([]string, 0, len(lines))
    
    // 保留开头的重要信息
    compressed = append(compressed, lines[:keepStart]...)
    
    // 压缩中间部分
    middleStart := keepStart
    middleEnd := len(lines) - keepStart
    if middleEnd > middleStart {
        compressed = append(compressed, fmt.Sprintf("...[compressed %d lines]...", middleEnd-middleStart))
    }
    
    // 保留结尾的最新数据
    compressed = append(compressed, lines[middleEnd:]...)
    
    return strings.Join(compressed, "\n")
}
```

### 并发AI调用

```go
// 并发AI模型调用
func BatchAIRequest(models []string, prompt string, timeout time.Duration) (map[string]string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    results := make(map[string]string)
    var wg sync.WaitGroup
    var mu sync.Mutex
    errChan := make(chan error, len(models))
    
    for _, model := range models {
        wg.Add(1)
        go func(m string) {
            defer wg.Done()
            
            result, err := callAIModel(ctx, m, prompt)
            if err != nil {
                errChan <- fmt.Errorf("model %s failed: %w", m, err)
                return
            }
            
            mu.Lock()
            results[m] = result
            mu.Unlock()
        }(model)
    }
    
    wg.Wait()
    close(errChan)
    
    // 检查错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return results, fmt.Errorf("some models failed: %v", errors)
    }
    
    return results, nil
}
```

## 前端性能优化

### React性能优化

```javascript
// 组件性能优化
import React, { memo, useMemo, useCallback } from 'react';

// 使用memo避免不必要的重渲染
const OptimizedComponent = memo(({ data, onUpdate }) => {
  // 使用useMemo缓存计算结果
  const processedData = useMemo(() => {
    return data.map(item => ({
      ...item,
      processed: expensiveCalculation(item.value)
    }));
  }, [data]);

  // 使用useCallback缓存函数
  const handleClick = useCallback((id) => {
    onUpdate(id);
  }, [onUpdate]);

  return (
    <div>
      {processedData.map(item => (
        <MemoizedItem 
          key={item.id} 
          item={item} 
          onClick={handleClick}
        />
      ))}
    </div>
  );
});

// 虚拟滚动实现
const VirtualizedList = ({ items, itemHeight, renderItem }) => {
  const [scrollTop, setScrollTop] = useState(0);
  const containerRef = useRef();
  
  const visibleCount = Math.ceil(containerRef.current?.clientHeight / itemHeight) || 0;
  const startIndex = Math.floor(scrollTop / itemHeight);
  const endIndex = Math.min(startIndex + visibleCount, items.length);
  
  const visibleItems = useMemo(() => {
    return items.slice(startIndex, endIndex).map((item, index) => ({
      ...item,
      index: startIndex + index
    }));
  }, [items, startIndex, endIndex]);
  
  const totalHeight = items.length * itemHeight;
  const offsetY = startIndex * itemHeight;
  
  return (
    <div 
      ref={containerRef}
      style={{ height: '100%', overflow: 'auto' }}
      onScroll={(e) => setScrollTop(e.target.scrollTop)}
    >
      <div style={{ height: totalHeight, position: 'relative' }}>
        <div style={{ 
          transform: `translateY(${offsetY}px)` 
        }}>
          {visibleItems.map(item => (
            <div key={item.id} style={{ height: itemHeight }}>
              {renderItem(item)}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
```

### 代码分割和懒加载

```javascript
// 路由级代码分割
import { lazy, Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const StrategyStudio = lazy(() => import('./pages/StrategyStudio'));
const Backtest = lazy(() => import('./pages/Backtest'));

function App() {
  return (
    <Router>
      <Suspense fallback={<div>Loading...</div>}>
        <Routes>
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/strategy" element={<StrategyStudio />} />
          <Route path="/backtest" element={<Backtest />} />
        </Routes>
      </Suspense>
    </Router>
  );
}
```

## 性能测试基准

### 基准测试配置

```go
// Go基准测试
func BenchmarkAPITraderCreate(b *testing.B) {
    // 准备测试数据
    setupTestData()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            resp, err := http.Post("/api/v1/traders", "application/json", 
                strings.NewReader(testTraderJSON))
            if err != nil {
                b.Fatal(err)
            }
            resp.Body.Close()
        }
    })
}

func BenchmarkDatabaseQuery(b *testing.B) {
    db := setupTestDB()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var traders []Trader
        db.Where("is_running = ?", true).Find(&traders)
    }
}
```

### 性能测试脚本

```bash
#!/bin/bash
# 性能测试脚本

# API性能测试
echo "=== API Performance Test ==="
wrk -t12 -c400 -d30s http://localhost:8888/api/v1/health

# 数据库性能测试
echo "=== Database Performance Test ==="
pgbench -c 50 -j 10 -T 300 -f test_script.sql nofx_db

# 前端性能测试
echo "=== Frontend Performance Test ==="
lighthouse http://localhost:3300 --output=json --output-path=perf-report.json

# AI模型性能测试
echo "=== AI Model Performance Test ==="
ab -n 100 -c 10 -p test_prompt.json -T application/json http://localhost:8888/api/v1/ai/generate
```

## 性能调优最佳实践

### 系统级优化

```bash
# Linux系统优化
# 调整文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 调整内核网络参数
cat >> /etc/sysctl.conf << EOF
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
EOF

sysctl -p
```

### 应用级优化清单

```markdown
## 性能优化检查清单

### 数据库优化
- [ ] 添加必要的数据库索引
- [ ] 优化慢查询SQL
- [ ] 配置合适的连接池大小
- [ ] 启用查询缓存
- [ ] 定期维护和优化表

### 缓存策略
- [ ] 实现多级缓存架构
- [ ] 配置合适的缓存过期时间
- [ ] 使用缓存预热机制
- [ ] 实现缓存失效策略
- [ ] 监控缓存命中率

### 并发优化
- [ ] 合理设置Goroutine池大小
- [ ] 使用批量处理减少I/O
- [ ] 实现连接复用
- [ ] 优化锁竞争
- [ ] 使用异步处理

### 内存管理
- [ ] 实现对象池复用
- [ ] 定期进行内存分析
- [ ] 优化数据结构
- [ ] 避免内存泄漏
- [ ] 配置合适的GC策略

### 网络优化
- [ ] 启用HTTP/2
- [ ] 配置连接池
- [ ] 启用Keep-Alive
- [ ] 压缩传输数据
- [ ] 使用CDN加速
```

## 性能监控告警

### Prometheus告警规则

```yaml
# prometheus/rules.yml
groups:
- name: nofx.rules
  rules:
  - alert: HighCPUUsage
    expr: rate(process_cpu_seconds_total[5m]) > 0.8
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High CPU usage detected"
      description: "CPU usage is above 80% for more than 2 minutes"

  - alert: HighMemoryUsage
    expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.85
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High memory usage"
      description: "Memory usage is above 85%"

  - alert: SlowAPIResponse
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Slow API response"
      description: "95th percentile API response time exceeds 500ms"

  - alert: DatabaseSlowQuery
    expr: rate(pg_stat_statements_mean_time[5m]) > 1000
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "Slow database queries"
      description: "Average query time exceeds 1000ms"
```

### 性能仪表板

```json
{
  "dashboard": {
    "title": "NOFX Performance Dashboard",
    "panels": [
      {
        "title": "System Resources",
        "type": "graph",
        "targets": [
          "node_cpu_seconds_total",
          "node_memory_MemAvailable_bytes",
          "node_disk_io_time_seconds_total"
        ]
      },
      {
        "title": "API Performance",
        "type": "graph",
        "targets": [
          "http_requests_total",
          "http_request_duration_seconds"
        ]
      },
      {
        "title": "Database Metrics",
        "type": "graph",
        "targets": [
          "pg_stat_statements_calls",
          "pg_stat_statements_total_time"
        ]
      }
    ]
  }
}
```

---
*性能优化文档版本: v1.0*
*最后更新: 2026年3月*