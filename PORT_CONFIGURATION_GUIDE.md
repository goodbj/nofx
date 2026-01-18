# 端口配置管理指南

## 概述
本指南介绍如何在开发版中使用 `.env.example` 文件进行端口配置管理。`.env.example` 是一个示例模板文件，用户应将其复制为 `.env` 文件进行实际配置。

## 开发版配置说明

### 环境变量定义
在 `.env.example` 文件中定义了以下开发版端口变量：

# Backend API server port
NOFX_BACKEND_PORT=8888

# Frontend web interface port
NOFX_FRONTEND_PORT=3300

# API server port
API_SERVER_PORT=${NOFX_BACKEND_PORT}

这些是开发版的关键端口配置变量：
- `NOFX_BACKEND_PORT=8888` - 后端API服务端口
- `NOFX_FRONTEND_PORT=3300` - 前端Web界面端口
- `API_SERVER_PORT` - API服务器端口，通常与后端端口相同

## 使用方法

### 1. 初始化配置
复制 `.env.example` 文件为 `.env`：
```bash
cp .env.example .env
```

或者，也可以使用 `.env.template` 作为更详细的模板：
```bash
cp .env.template .env
```

此 `.env` 文件将作为实际运行时的配置文件，包含所有环境变量设置。

在 `.env` 文件中，您可以：
- 通过修改 `NOFX_BACKEND_PORT` 来更改后端API端口
- 通过修改 `NOFX_FRONTEND_PORT` 来更改前端端口
- API_SERVER_PORT 会自动跟随后端端口的变化

### 2. 切换到开发版
当前默认设置为开发版，无需额外操作。

### 3. 切换到稳定版
编辑 `.env` 文件，将当前激活端口部分修改为：
```bash
# Or for stable version:
API_SERVER_PORT=${STABLE_BACKEND_PORT}
NOFX_BACKEND_PORT=${STABLE_BACKEND_PORT}
NOFX_FRONTEND_PORT=${STABLE_FRONTEND_PORT}
```

并注释掉开发版的设置：
```bash
# Current active ports (currently set to development)
# API_SERVER_PORT=${DEV_BACKEND_PORT}
# NOFX_BACKEND_PORT=${DEV_BACKEND_PORT}
# NOFX_FRONTEND_PORT=${DEV_FRONTEND_PORT}
```

### 4. 自定义端口
如需自定义端口，只需修改 `.env` 文件中的相应变量值：
1. 修改 `DEV_BACKEND_PORT`、`DEV_FRONTEND_PORT`、`STABLE_BACKEND_PORT` 或 `STABLE_FRONTEND_PORT` 的值
2. 重启 Docker 服务使更改生效

## Docker 配置集成

要使 Docker 配置文件使用这些环境变量，需要在 `docker-compose.dev.watch.yml` 和 `docker-compose.stable.yml` 中引用这些变量：

### 开发版配置示例
```yaml
services:
  nofx-dev-watch:
    ports:
      - "${NOFX_BACKEND_PORT:-8888}:${NOFX_BACKEND_PORT:-8888}"
    environment:
      - NOFX_BACKEND_PORT=${NOFX_BACKEND_PORT:-8888}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:${NOFX_BACKEND_PORT:-8888}/api/health"]
  nofx-frontend-dev-watch:
    ports:
      - "${NOFX_FRONTEND_PORT:-3300}:5173"
```

### 稳定版配置示例
```yaml
services:
  nofx:
    ports:
      - "${NOFX_BACKEND_PORT:-8888}:8888"
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:${NOFX_BACKEND_PORT:-8888}/api/health"]
  nofx-frontend:
    ports:
      - "${NOFX_FRONTEND_PORT:-3300}:80"
```

## 注意事项
- 修改端口后需要重启对应的 Docker 服务
- 确保新端口未被其他服务占用
- 开发版和稳定版的端口应保持不同以避免冲突
- 确保 Docker Compose 文件正确引用了环境变量