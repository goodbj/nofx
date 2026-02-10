# NOFX 开发版 Docker 部署指南

## 概述

本指南介绍了如何使用 Docker 部署 NOFX 开发版系统，包括前端和后端服务。该系统配置为开发模式，支持热更新和 Chrome 无头模式。

## 系统架构

- **前端**: nofx-dev-frontend (端口 3300)
- **后端**: nofx-dev-backend (端口 8888)
- **功能**: Chrome 无头模式（不弹窗）
- **Chrome 用户数据存储目录**: E:\AI\nofx_Dev\storage
- **数据库目录**: E:\AI\nofx_Dev\data\data.db

## 先决条件

在开始部署之前，请确保已安装以下软件：

1. **Docker Desktop** (推荐最新版本)
2. **Docker Compose** (通常随 Docker Desktop 一起安装)

## 部署步骤

### 方法一：使用批处理脚本（推荐）

1. 运行启动脚本：
   ```
   start_nofx_dev_docker.bat
   ```

2. 脚本将自动：
   - 验证 Docker 和 Docker Compose 是否可用
   - 构建 Docker 镜像
   - 启动前端和后端服务
   - 显示服务状态

### 方法二：手动部署

1. 打开终端或命令提示符，导航到项目根目录
2. 运行以下命令启动服务：
   ```
   docker compose -f docker-compose.nofx-dev.yml up -d
   ```

## 服务访问

部署完成后，可以通过以下地址访问服务：

- **前端界面**: http://localhost:3300
- **后端 API**: http://localhost:8888

## 功能特性

### Chrome 无头模式
- Chrome 浏览器在后台以无头模式运行
- 不会弹出任何窗口
- 用户数据存储在 `E:\AI\nofx_Dev\storage` 目录

### 热更新
- 后端代码修改后会自动重新加载
- 前端代码修改后会自动刷新页面

### 数据持久化
- 数据库文件位于 `E:\AI\nofx_Dev\data\data.db`
- Chrome 用户数据持久保存在 `E:\AI\nofx_Dev\storage`

## 管理命令

### 查看服务状态
```
docker compose -f docker-compose.nofx-dev.yml ps
```

### 查看日志
```
docker compose -f docker-compose.nofx-dev.yml logs -f
```

### 停止服务
```
docker compose -f docker-compose.nofx-dev.yml down
```

### 重新构建并启动
```
docker compose -f docker-compose.nofx-dev.yml up --build -d
```

## 批处理脚本

项目提供了以下便利的批处理脚本：

- `start_nofx_dev_docker.bat` - 启动服务
- `stop_nofx_dev_docker.bat` - 停止服务
- `restart_nofx_dev_docker.bat` - 重启服务

## 故障排除

### Docker 未运行
如果遇到 "Docker is not running" 错误，请先启动 Docker Desktop。

### 端口冲突
如果遇到端口冲突，请检查是否有其他服务占用了 3300 或 8888 端口。

### 权限问题
确保有足够的权限访问项目目录和数据目录。

## 目录结构

```
E:\AI\nofx_Dev\
├── docker-compose.nofx-dev.yml (Docker Compose 配置文件)
├── docker/
│   ├── Dockerfile.backend.dev.watch (后端 Dockerfile)
│   └── Dockerfile.frontend.dev.watch (前端 Dockerfile)
├── storage/ (Chrome 用户数据存储)
└── data/ (数据库文件存储)
    └── data.db
```

## 自定义配置

如需自定义配置，可以编辑 `docker-compose.nofx-dev.yml` 文件：

- 修改端口映射
- 更改环境变量
- 调整卷挂载路径