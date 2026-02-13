# 统一端口配置

本文档记录项目中所有服务的端口配置，以避免端口冲突和硬编码。

## 服务端口分配

| 服务 | 端口 | 用途 | 环境变量 |
|------|------|------|----------|
| API Server | 8888 | 主API服务 | NOFX_BACKEND_PORT |
| Binance Proxy | 8081 | 币安API代理服务 | BINANCE_PROXY_PORT |
| Frontend | 3300 | 前端服务 | NOFX_FRONTEND_PORT |
| Profiling | 6060 | 性能分析 | 无 |

## 环境变量配置

在 `.env` 文件中应包含以下配置：

```bash
# API Server 端口
NOFX_BACKEND_PORT=8888

# Binance 代理服务端口
BINANCE_PROXY_PORT=8081

# 前端端口
NOFX_FRONTEND_PORT=3300
```

## 服务配置文件位置

- API 服务: `api/server.go`
- 代理服务: `api/binance_proxy/main.go`
- 代理服务配置: `api/binance_proxy/config.go`
- Docker 配置: `proxy_service/docker-compose.proxy.yml`
- Docker环境变量: `proxy_service/.env`
- Dockerfile: `api/binance_proxy/Dockerfile.simple`