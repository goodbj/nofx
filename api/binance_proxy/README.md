# 币安代理服务

币安代理服务是一个独立的HTTP代理，用于转发币安期货API请求，绕过地区限制。

## 配置

代理服务通过环境变量进行配置：

### 环境变量

- `PORT`: 代理服务监听端口（默认: 8081）
- `DEFAULT_TARGET_API_URL`: 默认目标API URL（默认: https://testnet.binancefuture.com）
- `DEBUG`: 调试模式（默认: false）

### 配置文件

可以使用 `.env` 文件来设置环境变量：

```bash
PORT=8081
DEFAULT_TARGET_API_URL=https://testnet.binancefuture.com
DEBUG=false
```

## 使用方法

### 启动服务

```bash
# 直接运行
go run *.go

# 或构建后运行
go build -o binance-proxy .
./binance-proxy
```

### API 使用

代理服务支持所有币安期货API端点，例如：

```
# 虚拟盘请求
GET http://localhost:8081/fapi/v1/time

# 实盘请求（通过请求头指定目标URL）
GET http://localhost:8081/fapi/v1/time
Header: X-Custom-API-URL: https://fapi.binance.com
```

## Docker 部署

```bash
# 构建并运行
docker build -t binance-proxy .
docker run -p 8081:8081 binance-proxy
```

## 环境变量示例

参见 `.env.example` 文件了解完整的配置选项。