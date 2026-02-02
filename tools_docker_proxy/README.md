# 透明代理服务 (Transparent Proxy Service)

## 概述

透明代理服务是一个独立的Docker容器，用于转发nofx到交易所API的请求。该服务旨在绕过地区限制，使nofx能够在受限地区访问交易所API。

## 功能特点

- **透明转发**：完全透明地转发HTTP请求，不对请求内容进行修改
- **目标URL替换**：从请求头 `X-Custom-API-URL` 获取真实目标地址
- **循环防护**：内置循环防护机制，防止请求被转发到自身
- **热更新支持**：支持配置热更新
- **Docker部署**：通过Docker容器化部署，易于管理和扩展

## 架构设计

```
nofx → 代理服务 → 交易所API
```

nofx将请求发送到代理服务，并在请求头 `X-Custom-API-URL` 中指定真实的目标URL。代理服务接收请求后，从头部获取真实目标URL，然后将请求转发到该地址。

## 部署

### Docker Compose 部署

```bash
# 构建并启动服务
docker compose up -d --build

# 或使用传统命令
docker-compose up -d --build
```

### 验证部署

```bash
# 检查服务状态
docker ps | grep tools_docker_proxy

# 检查健康状态
curl http://localhost:8081/health
```

## 配置

### 环境变量

- `PORT`: 代理服务监听端口 (默认: 8081)

### Docker Compose 配置

```yaml
services:
  tools_docker_proxy:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: tools_docker_proxy  # 容器名称
    ports:
      - "${PROXY_PORT:-8081}:8081"      # 端口映射
    environment:
      - PORT=8081
      - GIN_MODE=release
    restart: unless-stopped
    networks:
      - nofx-proxy-network
```

## 工作流程

1. nofx准备API请求，将真实交易所URL放入 `X-Custom-API-URL` 请求头
2. 请求发送到代理服务 (`http://localhost:8081`)
3. 代理服务从 `X-Custom-API-URL` 头部提取真实目标URL
4. 代理服务将请求路径和查询参数附加到真实URL
5. 代理服务将请求转发到真实的目标交易所API
6. 代理服务将响应返回给nofx

## 安全特性

- **循环防护**：防止请求被转发到自身，避免无限循环
- **请求头验证**：确保 `X-Custom-API-URL` 头部存在
- **URL验证**：验证目标URL格式

## 日志

代理服务会输出详细的请求处理日志，包括：
- 接收的请求信息
- 目标URL解析结果
- 请求转发状态
- 错误信息（如有）

## 故障排除

### 服务未启动

检查Docker服务状态：
```bash
docker ps -a
```

查看容器日志：
```bash
docker logs tools_docker_proxy
```

### 请求被阻止

检查循环防护日志：
- 确保nofx不直接连接到代理服务的端口
- 验证 `X-Custom-API-URL` 头部值正确

### 连接超时

- 验证代理服务在8081端口运行
- 检查防火墙设置
- 确认nofx配置正确

## 维护

### 更新服务

```bash
# 停止当前服务
docker compose down

# 拉取最新代码（如有）
git pull

# 重新构建并启动
docker compose up -d --build
```

### 监控

定期检查容器日志：
```bash
docker logs tools_docker_proxy --tail 50
```

## API

### 健康检查

```
GET /health
```

返回：`OK`

### 通用代理端点

```
ANY /*path
```

将请求转发到 `X-Custom-API-URL` 头部指定的地址