# 端口配置更改说明

## 概述
本项目的所有端口配置现已统一管理在 `.env` 文件中。只需修改该文件中的端口变量，整个项目的端口配置将会统一更新。

## 端口配置

在 `.env` 文件中，您可以看到以下端口配置：

```bash
# Backend API server port
NOFX_BACKEND_PORT=8888

# Frontend web interface port  
NOFX_FRONTEND_PORT=3300

# API server port
API_SERVER_PORT=${NOFX_BACKEND_PORT}
```

## 如何更改端口

### 1. 修改后端端口
将 `NOFX_BACKEND_PORT` 的值更改为所需端口：
```bash
NOFX_BACKEND_PORT=9999
```

### 2. 修改前端端口
将 `NOFX_FRONTEND_PORT` 的值更改为所需端口：
```bash
NOFX_FRONTEND_PORT=4444
```

### 3. 重启服务
修改端口后，需要重启 Docker 服务以使更改生效：
```bash
docker compose -f docker-compose.dev.watch.yml down
docker compose -f docker-compose.dev.watch.yml up -d
```

## 影响范围

修改 `.env` 文件中的端口变量会影响以下组件：

- **后端服务**：API 服务器将在新的后端端口上运行
- **Docker 容器映射**：Docker 容器将映射到新端口
- **健康检查**：Docker 健康检查将检查新端口
- **前端代理**：前端开发服务器将代理 API 请求到新后端端口
- **环境变量**：所有服务都将使用更新后的端口值

## 注意事项

- 确保新端口未被其他服务占用
- 修改端口后必须重启 Docker 服务才能生效
- 前端访问地址也将相应改变（例如：如果前端端口改为 4444，则访问地址为 `http://localhost:4444`）
- 如果使用 Docker 部署，确保防火墙允许新端口的流量