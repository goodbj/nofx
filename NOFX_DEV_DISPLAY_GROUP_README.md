# NOFX 开发版显示分组服务 (nofx_dev_display)

## 概述

`nofx_dev_display` 是专门为开发调试设计的Docker服务分组，包含前端和后端服务，具有以下特点：

- **标签分组**: 服务统一归类到 `nofx_dev_display` 分组标签下
- **浏览器显示**: 后端启用Chrome浏览器显示功能（非无头模式）
- **热更新支持**: 前后端代码修改后自动重新加载
- **网络隔离**: 服务运行在独立的bridge网络中
- **数据持久化**: 用户数据和数据库持久化到本地目录

## 服务架构

```
nofx_dev_display 分组
├── 后端服务 (backend)
│   ├── 容器名: nofx-dev-backend-display
│   ├── 端口: 8888
│   ├── 功能: API服务 + Chrome浏览器显示
│   └── 网络: nofx-dev-display-network
│
└── 前端服务 (frontend)
    ├── 容器名: nofx-dev-frontend-display
    ├── 端口: 3300
    ├── 功能: Web界面 + 热更新
    └── 网络: nofx-dev-display-network
```

## 配置文件

- **主配置**: `docker-compose.nofx-dev-display.yml`
- **后端Dockerfile**: `docker/Dockerfile.backend.dev.display`
- **前端Dockerfile**: `docker/Dockerfile.frontend.dev.watch`

## 部署方式

### 1. 使用管理脚本（推荐）

```bash
# 启动服务
start_nofx_dev_display.bat

# 停止服务
stop_nofx_dev_display.bat

# 重启服务
restart_nofx_dev_display.bat
```

### 2. 直接使用Docker Compose

```bash
# 启动服务
docker compose -f docker-compose.nofx-dev-display.yml up -d

# 查看服务状态
docker compose -f docker-compose.nofx-dev-display.yml ps

# 查看日志
docker compose -f docker-compose.nofx-dev-display.yml logs -f

# 停止服务
docker compose -f docker-compose.nofx-dev-display.yml down
```

## 服务特性

### 后端服务 (nofx-dev-backend-display)

- **Chrome显示**: `CHROME_HEADLESS=false` 启用浏览器界面
- **热更新**: 源代码目录挂载实现代码变更自动重载
- **数据持久化**: 
  - Chrome用户数据: `./storage`
  - 数据库文件: `./data`
- **环境变量**:
  - `CHROME_HEADLESS=false` - 启用浏览器显示
  - `DISPLAY=:0` - X11显示配置
  - `CHROME_BIN=/usr/bin/chromium-browser`

### 前端服务 (nofx-dev-frontend-display)

- **热更新**: Vite开发服务器，代码变更实时刷新
- **端口映射**: 3300 → 5173
- **API代理**: 自动代理到后端服务
- **环境变量**:
  - `VITE_API_TARGET=http://backend:8888` - 后端服务地址

## 访问地址

- **前端界面**: http://localhost:3300
- **后端API**: http://localhost:8888
- **健康检查**: http://localhost:8888/health

## 目录结构

```
项目根目录/
├── docker-compose.nofx-dev-display.yml  # 分组服务配置
├── start_nofx_dev_display.bat          # 启动脚本
├── stop_nofx_dev_display.bat           # 停止脚本
├── restart_nofx_dev_display.bat        # 重启脚本
├── storage/                            # Chrome用户数据持久化
├── data/                               # 数据库文件持久化
└── web/                                # 前端源代码
```

## 网络配置

- **网络名称**: `nofx-dev-display-network`
- **网络类型**: bridge
- **服务间通信**: 通过服务名 `backend` 和 `frontend` 直接访问
- **外部访问**: 通过端口映射访问

## 数据持久化

### 存储目录

1. **Chrome用户数据**: `./storage`
   - 浏览器配置文件
   - 缓存数据
   - 用户偏好设置

2. **数据库文件**: `./data`
   - SQLite数据库文件
   - 交易数据
   - 配置信息

### 持久化配置

```yaml
volumes:
  - ./storage:/app/storage    # Chrome数据
  - ./data:/app/data          # 数据库文件
```

## 开发调试

### 热更新机制

- **后端**: Go源代码文件变更时自动重启
- **前端**: Vite HMR (Hot Module Replacement) 实时更新

### 日志查看

```bash
# 查看所有服务日志
docker compose -f docker-compose.nofx-dev-display.yml logs -f

# 查看特定服务日志
docker compose -f docker-compose.nofx-dev-display.yml logs -f backend
docker compose -f docker-compose.nofx-dev-display.yml logs -f frontend
```

### 进入容器调试

```bash
# 进入后端容器
docker exec -it nofx-dev-backend-display /bin/bash

# 进入前端容器
docker exec -it nofx-dev-frontend-display /bin/sh
```

## 常见问题

### 1. 端口冲突

**问题**: 端口3300或8888已被占用

**解决**: 
```bash
# 查看端口占用
netstat -ano | findstr :3300
netstat -ano | findstr :8888

# 停止占用进程或修改端口配置
```

### 2. 浏览器无法显示

**问题**: Chrome浏览器在容器中无法显示

**解决**:
- 确认 `CHROME_HEADLESS=false` 环境变量已设置
- Windows环境下需要配置X11服务器（如VcXsrv）
- 检查 `/tmp/.X11-unix` 挂载配置

### 3. 前后端通信失败

**问题**: 前端无法连接到后端API

**解决**:
- 确认后端服务已正常启动
- 检查网络配置 `nofx-dev-display-network`
- 验证 `VITE_API_TARGET` 环境变量配置

### 4. 数据持久化问题

**问题**: 重启后数据丢失

**解决**:
- 确认 `./storage` 和 `./data` 目录存在
- 检查volume挂载配置
- 验证目录权限设置

## 使用场景

此分组服务特别适用于:

1. **开发调试**: 需要观察浏览器行为的开发场景
2. **UI测试**: 前端界面与后端交互的测试
3. **演示展示**: 需要展示完整应用运行效果
4. **教学培训**: 展示完整的应用运行流程
5. **问题排查**: 调试浏览器相关的问题

## 与其它配置的区别

| 配置名称 | 浏览器模式 | 主要用途 | 网络配置 |
|---------|-----------|----------|----------|
| nofx_dev_display | 显示模式 | 开发调试 | 独立网络 |
| nofx-dev | 无头模式 | 生产部署 | 独立网络 |
| nofx-dev-grouped | 无头模式 | 开发联调 | 共享网络 |

## 维护建议

1. **定期清理**: 定期清理不必要的容器和镜像
2. **数据备份**: 重要数据定期备份 `./storage` 和 `./data` 目录
3. **日志监控**: 定期检查服务日志，及时发现潜在问题
4. **资源监控**: 监控容器资源使用情况，避免资源耗尽