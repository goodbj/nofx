# 币安数据获取插件部署和使用指南

本指南介绍如何部署和使用币安数据获取插件，该插件专门负责绕过币安监管，从主程序获取连接参数，专门负责获取数据、发送数据和执行命令。

## 目录结构

```
nofx_Dev/
├── api/
│   └── binance_proxy/          # 币安代理服务
│       ├── main.go             # 服务启动入口
│       ├── service.go          # 核心业务逻辑
│       ├── config.go           # 配置管理
│       └── Dockerfile          # Docker镜像构建
├── dataprovider/               # 数据提供者模块
│   └── provider.go             # 数据提供者实现
├── docker-compose.proxy.yml    # 代理服务部署配置
└── .env                        # 环境变量配置
```

## 部署步骤

### 1. 构建代理服务

#### 方法一：直接运行
```bash
# 进入代理服务目录
cd api/binance_proxy

# 安装依赖
go mod tidy

# 运行代理服务
go run main.go
```

#### 方法二：Docker部署
```bash
# 构建并启动代理服务
docker-compose -f docker-compose.proxy.yml up --build
```

### 2. 配置环境变量

在 `.env` 文件中设置以下参数：

```
# 启用币安代理模式
USE_BINANCE_PROXY=true

# 代理服务地址
BINANCE_PROXY_URL=http://localhost:8081
```

### 3. 启动主程序

```bash
# 返回项目根目录
cd ../..

# 启动主程序
go run main.go
```

## 工作原理

### 代理模式 (USE_BINANCE_PROXY=true)
1. 主程序接收请求
2. 通过数据提供者接口调用相应方法
3. ProxyDataProvider构造HTTP请求发送到代理服务
4. 代理服务调用币安API获取数据
5. 代理服务返回数据给主程序
6. 主程序返回结果给前端

### 直接模式 (USE_BINANCE_PROXY=false)
1. 主程序接收请求
2. 通过数据提供者接口调用相应方法
3. DirectDataProvider直接调用币安API
4. 主程序返回结果给前端

## API端点

代理服务提供以下API端点：

- `GET /health` - 健康检查
- `GET /api/proxy/balance` - 获取账户余额
- `GET /api/proxy/positions` - 获取持仓信息
- `GET /api/proxy/klines` - 获取K线数据
- `GET /api/proxy/account` - 获取账户信息
- `GET /api/proxy/trades` - 获取交易历史
- `POST /api/proxy/orders` - 下单
- `DELETE /api/proxy/orders` - 撤单
- `GET /api/proxy/orders` - 获取订单

## 安全特性

1. **API密钥隔离** - 在代理模式下，API密钥通过HTTP请求头传递，不会在主程序中直接存储
2. **环境变量控制** - 通过环境变量控制是否启用代理，可随时切换模式
3. **CORS配置** - 代理服务配置允许跨域访问，确保前端可以访问

## 切换模式

### 启用代理模式
```
USE_BINANCE_PROXY=true
BINANCE_PROXY_URL=http://localhost:8081
```

### 禁用代理模式（直接调用）
```
USE_BINANCE_PROXY=false
```

修改环境变量后需要重启主程序。

## 故障排除

### 1. 检查代理服务是否运行
```bash
curl http://localhost:8081/health
```

### 2. 检查端口占用
确保8081端口未被其他服务占用。

### 3. 查看日志
检查代理服务和主程序的日志输出。

## 优势

1. **监管绕过** - 通过独立的服务处理币安API调用，可以更容易地应对监管变化
2. **安全隔离** - API密钥和其他敏感信息在独立环境中处理
3. **灵活部署** - 可以在不同地理位置部署代理服务以绕过地域限制
4. **可扩展性** - 可以轻松添加更多币安API功能
5. **向后兼容** - 不影响现有功能，可平滑切换