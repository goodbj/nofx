# 透明代理安装和配置指南

## 概述

透明代理功能允许nofx通过代理服务访问受限地区的交易所API，绕过地区限制。该功能通过Docker容器部署，支持热更新。

## 安装步骤

### 1. 环境准备

确保系统已安装以下软件：
- Docker Desktop 或 Docker Engine
- Git
- Go 1.19+ (可选，用于本地构建)

### 2. 配置代理服务

代理服务位于 `tools_docker_proxy` 目录，包含以下文件：
- `Dockerfile`: Docker镜像构建文件
- `docker-compose.yml`: 容器编排配置
- `main.go`: 代理服务核心实现
- `go.mod` / `go.sum`: Go模块依赖

### 3. 启动代理服务

在 `tools_docker_proxy` 目录下执行：

```bash
docker compose up -d --build
```

或使用传统命令：

```bash
docker-compose up -d --build
```

### 4. 验证代理服务

确认代理服务运行在8081端口：

```bash
curl http://localhost:8081/health
```

应返回 "OK" 响应。

## 配置nofx使用代理

### 1. 环境变量配置

在 `.env` 文件中配置以下参数：

```bash
# 启用Binance代理功能
USE_BINANCE_PROXY=true

# 代理服务端口
BINANCE_PROXY_PORT=8081

# 代理服务URL
BINANCE_PROXY_URL=http://localhost:8081
```

### 2. 重启nofx服务

修改配置后，重启nofx后端服务以使配置生效。

## 工作原理

1. nofx内部使用真实的交易所URL进行业务逻辑处理
2. 通过 `CustomTransport` 将真实交易所URL放入 `X-Custom-API-URL` 请求头
3. 请求发送到代理服务 (`http://localhost:8081`)
4. 代理服务从请求头获取真实目标URL
5. 代理服务将请求路径和参数附加到真实URL并转发到交易所
6. 完成对交易所API的安全访问

## 故障排除

### 代理服务未启动

检查Docker服务是否运行：
```bash
docker ps
```

查看代理服务日志：
```bash
docker logs tools_docker_proxy
```

### 连接失败

确认以下配置：
- 代理服务在8081端口运行
- `.env` 文件中代理配置正确
- nofx服务已重启

### 循环防护机制

代理服务具有自循环防护机制，防止请求被转发到自身。

## 维护

### 更新代理服务

修改代理服务代码后，重新构建Docker镜像：

```bash
docker compose down
docker compose up -d --build
```

### 监控

代理服务会在控制台输出请求处理日志，便于监控和调试。