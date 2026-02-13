# NOFX Docker 开发环境热重载使用手册

## 概述

本手册介绍了如何使用Docker在开发模式下运行NOFX项目，实现代码修改后即时预览效果。Docker开发环境提供了与生产环境一致的网络性能和系统配置，同时支持代码热重载功能。

## 目录

1. [环境要求](#环境要求)
2. [开发模式配置](#开发模式配置)
3. [启动开发环境](#启动开发环境)
4. [开发工作流程](#开发工作流程)
5. [日志监控](#日志监控)
6. [停止开发环境](#停止开发环境)
7. [故障排除](#故障排除)
8. [注意事项](#注意事项)

## 环境要求

- Docker Desktop 或 Docker Engine
- Docker Compose (v2.0+)
- 至少 4GB 可用内存
- 项目根目录的 `.env` 文件已正确配置

## 开发模式配置

本项目提供了两种Docker开发模式配置：

### 1. 标准开发模式（无热重载）

使用 `docker-compose.dev.yml`，适用于：
- 部署测试
- 稳定运行
- 性能测试

### 2. 热重载开发模式（推荐用于日常开发）

使用 `docker-compose.dev.watch.yml`，适用于：
- 实时代码修改预览
- 快速迭代开发
- 即时错误反馈

## 启动开发环境

### 热重载开发模式

```bash
# 进入项目目录
cd e:\AI\nofx_Dev

# 构建并启动开发环境（带热重载）
docker-compose -f docker-compose.dev.watch.yml up -d --build

# 或者，如果您只想启动而不后台运行（便于查看实时日志）
docker-compose -f docker-compose.dev.watch.yml up --build
```

### 标准开发模式

```bash
# 进入项目目录
cd e:\AI\nofx_Dev

# 构建并启动标准开发环境
docker-compose -f docker-compose.dev.yml up -d --build
```

## 开发工作流程

### 后端开发（Go语言）

1. **修改Go源代码**：
   - 在本地编辑器中修改项目根目录下的Go文件
   - 保存文件后，Air工具会自动检测到变化

2. **实时反馈**：
   - Air自动重新编译Go代码
   - 服务自动重启（通常在几秒内完成）
   - 您的更改立即生效

3. **API测试**：
   - 访问 `http://localhost:8888` 测试API变更
   - 通过前端界面验证后端逻辑

### 前端开发（React/Vue等）

1. **修改前端源代码**：
   - 修改 `web/` 目录下的前端文件
   - 保存后，前端开发服务器会检测到变化

2. **实时预览**：
   - 访问 `http://localhost:3300` 查看前端变更
   - 浏览器会自动刷新或热更新

### 验证变更

- **API端点**：`http://localhost:8888/api/health`
- **前端界面**：`http://localhost:3300`
- **后端服务状态**：`http://localhost:8888/api/status`

## 日志监控

### 查看实时日志

```bash
# 查看后端日志（热重载模式）
docker logs -f nofx-trading-dev-watch

# 查看前端日志
docker logs -f nofx-frontend-dev-watch

# 查看后端日志（标准模式）
docker logs -f nofx-trading-dev

# 查看前端日志（标准模式）
docker logs -f nofx-frontend-dev
```

### 分析日志

- **成功构建**：看到 "System started successfully" 表示后端服务启动成功
- **API可用**：看到 "API server starting at http://localhost:8888" 表示API服务可用
- **热重载触发**：看到 "build success" 表示代码重载成功

## 停止开发环境

### 停止单个环境

```bash
# 停止热重载开发环境
docker-compose -f docker-compose.dev.watch.yml down

# 停止标准开发环境
docker-compose -f docker-compose.dev.yml down
```

### 停止所有开发环境

```bash
# 停止所有相关的开发容器
docker-compose -f docker-compose.dev.watch.yml down
docker-compose -f docker-compose.dev.yml down

# 清理未使用的Docker资源（可选）
docker system prune -f
```

## 故障排除

### 常见问题及解决方案

#### 1. 端口被占用
**症状**：`port is already allocated` 错误
**解决方案**：
```bash
# 检查占用端口的进程
netstat -ano | findstr :8888
netstat -ano | findstr :3300

# 终止占用端口的进程
taskkill /pid <PID> /f
```

#### 2. 代码更改未生效
**症状**：修改代码后，容器未自动重启
**解决方案**：
- 确保文件保存正确
- 检查Air工具是否正常运行
- 重新启动开发环境

#### 3. 依赖问题
**症状**：构建失败，出现包缺失错误
**解决方案**：
```bash
# 清理缓存并重新构建
docker-compose -f docker-compose.dev.watch.yml down
docker-compose -f docker-compose.dev.watch.yml build --no-cache
docker-compose -f docker-compose.dev.watch.yml up -d
```

#### 4. 网络连接问题
**症状**：无法访问 `http://localhost:8888` 或 `http://localhost:3300`
**解决方案**：
```bash
# 检查容器状态
docker ps

# 检查容器网络
docker inspect <container-name>
```

### 调试技巧

1. **查看容器内部**：
   ```bash
   # 进入后端容器
   docker exec -it nofx-trading-dev-watch sh
   
   # 进入前端容器
   docker exec -it nofx-frontend-dev-watch sh
   ```

2. **检查挂载卷**：
   ```bash
   # 验证代码是否正确挂载
   docker exec -it nofx-trading-dev-watch ls -la /app
   ```

## 注意事项

### 性能影响

- **热重载模式**：每次保存都会触发重建，可能影响性能，但对开发友好
- **资源消耗**：Docker容器会占用额外内存和CPU资源
- **磁盘空间**：Docker镜像和容器会占用磁盘空间

### 最佳实践

1. **定期清理**：
   - 定期清理未使用的Docker镜像和容器
   - 使用 `docker system prune` 释放空间

2. **分支开发**：
   - 在不同Git分支上开发时，建议清理旧的容器和镜像
   - 使用不同的Compose文件或环境变量区分开发场景

3. **依赖管理**：
   - 添加新的Go依赖后，需要重新构建镜像
   - 修改 `go.mod` 或 `go.sum` 后执行 `--build` 参数

4. **配置管理**：
   - 确保 `.env` 文件包含所有必要的环境变量
   - 不同环境使用不同的 `.env` 文件

### 安全考虑

- **敏感信息**：确保 `.env` 文件不包含在Docker镜像中
- **权限设置**：Docker容器以适当权限运行
- **网络隔离**：使用Docker网络隔离开发环境

## 扩展配置

### 自定义构建参数

如需修改构建参数，可以编辑 `docker-compose.dev.watch.yml` 文件：

- 修改端口映射
- 调整环境变量
- 更改挂载卷路径
- 调整容器资源限制

### 多环境支持

可根据需要创建更多Compose文件：

- `docker-compose.staging.yml` - 预发布环境
- `docker-compose.test.yml` - 测试环境
- `docker-compose.debug.yml` - 调试环境

### 便捷脚本使用

为了简化日常开发操作，我们提供了几个便捷的批处理脚本：

### 当前运行状态

您可以通过以下命令检查当前运行的容器：

```
docker ps
```

在典型的开发环境中，您会看到类似以下的输出：

- `nofx-trading-dev-watch` - 后端服务，监听端口8888
- `nofx-frontend-dev-watch` - 前端服务，监听端口3300

### 服务健康检查

可以使用以下URL检查服务状态：

- 后端API健康检查：http://localhost:8888/api/health
- 前端界面访问：http://localhost:3300

### 开发环境验证

要验证开发环境是否正常工作，请执行以下操作：

1. 访问API健康检查端点确认后端运行正常
2. 访问前端页面确认UI加载正常
3. 修改源代码中的任意文件，观察容器是否自动重启

### 数据持久化配置

为了解决重启服务后数据丢失的问题，我们配置了数据持久化：

1. **数据库配置**：在 `.env` 文件中将 `DB_PATH` 设置为 `data/data.db`
2. **数据卷挂载**：Docker Compose 配置中挂载了 `./data:/app/data` 卷
3. **持久化存储**：数据库文件将保存在项目根目录的 `data/` 文件夹中

现在，即使重启容器，数据也会保留在 `data/data.db` 文件中。

### 前端热重载功能

前端现在使用Vite开发服务器，支持真正的热重载功能：

1. **修改前端代码**：编辑`web/src`目录下的任何文件
2. **实时预览**：在浏览器中访问 http://localhost:3300，更改将自动显示
3. **Vite控制台**：查看前端容器日志以了解热重载状态

```bash
docker logs -f nofx-frontend-dev-watch
```

前端代码修改将立即反映在浏览器中，无需重新构建或重启容器。

1. **start_dev_watch.bat** - 启动热重载开发环境
2. **stop_dev_watch.bat** - 停止开发环境
3. **restart_dev_watch.bat** - 重启开发环境

这些脚本会自动处理目录切换和Docker Compose命令，使您能够快速启动开发环境。

---

**提示**：使用Docker开发环境不仅能提供与生产环境一致的运行时条件，还能解决网络连接问题，提升开发效率。建议在日常开发中使用此配置。