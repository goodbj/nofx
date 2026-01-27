# Guardian Docker热更新环境使用说明

## 概述

此文档介绍如何使用Guardian监测程序的Docker热更新环境。该环境支持代码更改后的实时更新，无需重新构建整个镜像。

## 环境配置

### 1. 热更新配置文件

- `docker-compose.dev.watch.minimal.yml` - 最小化热更新配置（仅包含后端和Guardian服务）
- `docker-compose.dev.with-guardian.watch.yml` - 完整热更新配置（包含前端、后端和Guardian服务）

### 2. 卷挂载配置

- `./guardian:/app` - Guardian代码目录热更新
- `./:/app` - 整个项目目录热更新（后端服务）

## 启动热更新环境

### 启动最小化环境（推荐用于开发）

```bash
cd e:\AI\nofx_Dev
docker-compose -f docker-compose.dev.watch.minimal.yml up -d --build
```

### 启动完整环境

```bash
cd e:\AI\nofx_Dev
docker-compose -f docker-compose.dev.with-guardian.watch.yml up -d --build
```

## 停止环境

```bash
cd e:\AI\nofx_Dev
docker-compose -f docker-compose.dev.watch.minimal.yml down
```

## 热更新工作原理

1. **代码自动同步**：通过Docker卷挂载，本地代码更改会自动同步到容器内
2. **实时生效**：对于Go程序，可能需要重启容器才能生效，或者使用支持热重载的工具
3. **开发效率**：无需每次修改代码后都重新构建镜像

## 开发流程

1. 启动热更新环境
2. 编辑本地 `guardian` 目录下的代码
3. （如需要）重启容器使更改生效
4. 验证更改效果

## 注意事项

1. **端口冲突**：确保端口8888（后端）和3300（前端，如启用）未被其他进程占用
2. **Docker资源**：热更新环境会持续监控文件变化，请确保Docker有足够的资源
3. **Chrome兼容性**：这是一个已发现的回归bug，原本在独立运行时Chrome初始化是成功的，但现在无论在Docker环境还是独立运行时都出现了Chrome初始化失败的问题。此问题现已通过优化Chrome启动参数得到解决：
   - 添加了适用于Docker和独立运行环境的兼容性参数
   - 包含了Windows环境的额外兼容性配置
   - 保持了对原有功能的支持

## 故障排除

### 常见问题

1. **端口被占用**：
   ```bash
   netstat -ano | findstr :8888
   taskkill /pid <PID> /f
   ```

2. **Docker服务未启动**：
   - 确保Docker Desktop已启动
   - 检查 `Get-Service com.docker.service` 状态

3. **Chrome浏览器初始化失败**：
   - 这是一个已知问题，无论在Docker环境还是独立运行时都可能出现
   - 不影响Guardian的核心监测功能，但浏览器自动化功能受限

## 构建镜像

如果需要重新构建镜像（当go.mod或依赖更改时）：

```bash
cd e:\AI\nofx_Dev
docker-compose -f docker-compose.dev.watch.minimal.yml build --no-cache
```

## 查看日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.dev.watch.minimal.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.dev.watch.minimal.yml logs -f nofx-guardian-dev
```

## 环境变量

热更新环境支持以下环境变量：

- `NOFX_BACKEND_PORT` - 后端端口（默认8888）
- `NOFX_FRONTEND_PORT` - 前端端口（默认3300）

## 性能优化建议

1. 仅挂载必要的目录以减少文件同步开销
2. 使用`.dockerignore`排除不必要的文件
3. 在开发完成后使用生产配置部署

---

**提示**：热更新环境主要用于开发和测试，在生产环境中请使用标准部署方式。