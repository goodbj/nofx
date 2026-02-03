# NOFX 开发版 Docker 部署帮助文档

## 概述

本文档提供 NOFX 开发版的 Docker 部署说明，适用于希望使用容器化方式运行开发环境的用户。该部署方案包括前端、后端服务，支持 Chrome 无头模式和热更新功能。

## 部署要求

- Docker Desktop 或 Docker Engine
- Docker Compose v2.x 或更高版本
- 至少 4GB 可用内存
- 端口 3300（前端）和 8888（后端）可用

## 部署文件

### 主要配置文件
- `docker-compose.nofx-dev-chrome-hot.yml` - Docker Compose 配置文件
- `start_nofx_dev_docker.bat` - 启动脚本
- `stop_nofx_dev_docker.bat` - 停止脚本

### 相关文档
- `DOCKER_DEPLOYMENT_README.md` - 详细部署文档
- `QUICK_START_DOCKER.md` - 快速启动指南

## 部署步骤

### 方法一：使用批处理脚本（推荐）

1. **启动服务**
   ```cmd
   start_nofx_dev_docker.bat
   ```

2. **停止服务**
   ```cmd
   stop_nofx_dev_docker.bat
   ```

### 方法二：使用 Docker Compose 命令

1. **启动服务**
   ```bash
   docker-compose -f docker-compose.nofx-dev-chrome-hot.yml up --build -d
   ```

2. **停止服务**
   ```bash
   docker-compose -f docker-compose.nofx-dev-chrome-hot.yml down
   ```

3. **查看日志**
   ```bash
   docker-compose -f docker-compose.nofx-dev-chrome-hot.yml logs -f
   ```

## 服务架构

### 后端服务 (nofx-dev-backend)
- **端口**: 8888
- **功能**: 
  - RESTful API 接口
  - Chrome 无头模式支持
  - 热更新功能（使用 Air 工具）
  - 数据库连接和管理
- **Dockerfile**: `docker/Dockerfile.backend.dev.watch`

### 前端服务 (nofx-dev-frontend)
- **端口**: 3300 (宿主机) → 3030 (容器内)
- **功能**: 
  - Web 用户界面
  - 与后端 API 通信
  - 热更新支持
- **Dockerfile**: `docker/Dockerfile.frontend.dev.watch`

## 特殊功能

### Chrome 无头模式
- 所有浏览器操作在后台运行，不会弹出窗口
- 适用于自动化测试和数据抓取
- 环境变量已预配置

### 热更新功能
- 代码更改时自动重新加载
- 缩短开发过程中的等待时间
- 支持 Go 和前端代码的实时更新

### 数据持久化
- SQLite 数据库文件存储在 `E:\AI\nofx_dev\data\data.db`
- 即使容器重启，数据也不会丢失

## 访问服务

- **前端界面**: http://localhost:3300
- **后端 API**: http://localhost:8888
- **API 健康检查**: http://localhost:8888/api/health

## 故障排除

### 1. 端口冲突
如果遇到端口冲突错误，请检查是否有其他服务占用了 3300 或 8888 端口。

### 2. 构建失败
- 确认 Docker 服务正在运行
- 检查网络连接是否正常
- 确保本地磁盘空间充足

### 3. 前端无法访问
- 确认后端服务已正常启动
- 检查防火墙设置

### 4. Chrome 无头模式异常
- 检查环境变量设置：`CHROME_HEADLESS=true`

## 维护建议

- 定期备份数据库文件 `data/data.db`
- 监控容器资源使用情况
- 根据需要调整容器资源配置

## 与其它版本的关系

此开发版 Docker 部署方案与以下版本保持一致的部署体系：
- 稳定版部署方案
- 生产版部署方案
- 其他开发环境部署方案

如需切换到其他版本，请参阅相应文档。