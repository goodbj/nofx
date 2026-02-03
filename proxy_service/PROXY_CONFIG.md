# 代理服务配置说明

## 服务概述
这是一个通用的HTTP代理服务，主要用于解决网络访问限制问题。特别适用于：
- 中国大陆网络环境访问海外API服务
- Docker容器内服务访问外网
- 任何需要网络代理转发的场景

## 目录结构
```
proxy_service/
├── binance_proxy/          # 代理服务源码
│   ├── main.go            # 主程序入口
│   ├── service.go         # 代理服务核心逻辑
│   ├── Dockerfile         # Docker构建文件
│   ├── go.mod             # Go模块依赖
│   └── go.sum             # Go依赖校验
├── PROXY_CONFIG.md        # 本配置说明文件
├── docker-compose.proxy.yml # Docker Compose配置
└── USAGE_GUIDE.md         # 使用指南
```

## 环境变量配置
- `USE_BINANCE_PROXY`: 全局开关，true启用代理，false禁用代理
- `BINANCE_PROXY_URL`: 代理服务地址，默认为 http://localhost:8082
- `PORT`: 代理服务监听端口，默认为 8082

## 部署方式
1. **Docker方式（推荐）**：使用 docker-compose.proxy.yml
2. **独立运行**：进入 binance_proxy 目录，执行 `go run main.go`
3. **编译运行**：执行 `go build` 后运行生成的二进制文件

## 适用场景
- Binance API访问代理
- 其他海外API访问代理
- 内网穿透
- Docker容器网络代理