# 代理服务使用指南

## 快速开始

### 1. Docker方式部署（推荐）
```bash
# 启动代理服务
docker-compose -f docker-compose.proxy.yml up -d

# 查看服务状态
docker-compose -f docker-compose.proxy.yml ps

# 查看日志
docker-compose -f docker-compose.proxy.yml logs -f
```

### 2. 本地开发模式
```bash
# 进入代理服务目录
cd proxy_service/binance_proxy

# 安装依赖
go mod tidy

# 运行服务
go run main.go
```

## 配置说明

### 环境变量
```bash
# 代理服务端口（默认8082）
PORT=8082

# 目标API基础URL（默认为币安期货API）
BINANCE_FUTURES_API_URL=https://fapi.binance.com
```

### Docker Compose 配置
```yaml
version: '3.8'

services:
  binance-proxy:
    build:
      context: ./proxy_service/binance_proxy
      dockerfile: Dockerfile
    container_name: nofx-binance-proxy
    ports:
      - "8082:8082"
    environment:
      - PORT=8082
      - BINANCE_FUTURES_API_URL=https://fapi.binance.com
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8082/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
```

## API端点

### 健康检查
- `GET /health` - 检查代理服务是否正常运行

### 代理端点
代理服务会转发所有请求到目标API，支持以下端点：
- `/fapi/v2/balance` - 余额查询
- `/fapi/v2/account` - 账户信息
- `/fapi/v2/positionRisk` - 持仓信息
- `/fapi/v1/order` - 订单操作
- `/fapi/v1/openOrders` - 未成交订单
- `/fapi/v1/allOrders` - 所有订单
- `/fapi/v1/ticker/price` - 价格查询
- `/fapi/v1/ticker/bookTicker` - 最优买卖档位
- `/fapi/v1/klines` - K线数据

## 故障排除

### 常见问题
1. **服务无法启动**
   - 检查端口是否被占用
   - 确认环境变量设置正确

2. **代理转发失败**
   - 检查目标API URL是否正确
   - 确认网络连接是否正常

3. **性能问题**
   - 检查并发连接数
   - 确认网络带宽是否足够

### 日志查看
代理服务会记录以下信息：
- 转发的请求方法和目标URL
- 请求处理时间
- 错误信息（如果有）

## 安全注意事项
- 代理服务会对敏感信息（如签名参数）进行脱敏处理
- 建议在生产环境中使用HTTPS
- 定期更新依赖包以修复安全漏洞