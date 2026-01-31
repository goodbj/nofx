# Binance Proxy Service

币安代理服务插件，专门负责绕过币安监管，从主程序获取连接参数，专门负责获取数据、发送数据和执行命令。

## 架构设计

### 核心组件

1. **BinanceProxyService** - 代理服务核心，负责与币安API通信
2. **DataProvider接口** - 定义数据提供者抽象接口
3. **DirectDataProvider** - 直接数据提供者，使用原有实现
4. **ProxyDataProvider** - 代理数据提供者，通过代理服务获取数据
5. **NewDataProvider工厂函数** - 根据环境变量返回相应的数据提供者

### 低耦合实现

- 通过`DataProvider`接口实现数据获取的抽象
- 主程序通过环境变量`USE_BINANCE_PROXY`控制是否启用代理模式
- 环境变量`BINANCE_PROXY_URL`指定代理服务地址

## 配置

### 环境变量

```bash
# 启用代理模式
USE_BINANCE_PROXY=true

# 代理服务地址
BINANCE_PROXY_URL=http://localhost:8081
```

### Docker部署

```bash
# 构建并启动代理服务
docker-compose -f docker-compose.proxy.yml up --build
```

## API端点

- `GET /health` - 健康检查
- `GET /api/proxy/balance` - 获取账户余额
- `GET /api/proxy/positions` - 获取持仓信息
- `GET /api/proxy/klines` - 获取K线数据
- `GET /api/proxy/account` - 获取账户信息
- `GET /api/proxy/trades` - 获取交易历史
- `POST /api/proxy/orders` - 下单
- `DELETE /api/proxy/orders` - 撤单
- `GET /api/proxy/orders` - 获取订单

## 使用方法

### 代理模式

当`USE_BINANCE_PROXY=true`时，主程序通过HTTP请求调用代理服务获取币安数据。

### 直接模式

当`USE_BINANCE_PROXY=false`时，主程序直接调用币安API（原有实现）。

## 优势

1. **解耦** - 主程序无需关心数据来源的具体实现
2. **灵活性** - 可随时切换数据获取方式
3. **可维护性** - 代理服务可以独立部署和扩展
4. **安全性** - 币安API密钥可以在隔离环境中处理