# nofx-dev-web 安装与部署指南

## 概述

本指南介绍如何安装和部署 nofx-dev-web（开发版-测试网页弹出功能版），该版本包含浏览器弹出功能，适用于测试浏览器自动化功能。

## 分组组成

nofx-dev-web 分组包含以下服务：

- **前端**: nofx-dev-frontend-display (端口 3300)
- **后端**: nofx-dev-backend-display (端口 8888)
- **功能**: Chrome非无头模式（会弹窗），用于测试浏览器自动化
- **文件目录**: E:\AI\nofx_dev
- **数据库目录**: E:\AI\nofx_dev\data\data.db（可读写）

## 前提条件

- Docker Desktop 已安装并运行
- Docker Compose 已安装
- 至少 4GB 可用磁盘空间
- 稳定的互联网连接

## 安装步骤

### 1. 克隆或下载项目

```bash
git clone <repository-url>
cd nofx_dev
```

### 2. 确保已有镜像

确保以下镜像已在本地存在：

- `setup-nofx-dev-frontend:latest`
- `nofx_dev-backend-display:latest`

如果不存在，请先构建这些镜像：

```bash
# 构建前端镜像
cd web
docker build -t setup-nofx-dev-frontend:latest .

# 构建后端镜像
cd ..
docker build -t nofx_dev-backend-display:latest -f docker/Dockerfile.backend.dev.display .
```

### 3. 启动 nofx-dev-web 分组

```bash
# 在项目根目录下运行
docker-compose -f docker-compose.dev.display.yml up -d
```

### 4. 验证部署

检查容器是否正在运行：

```bash
docker-compose -f docker-compose.dev.display.yml ps
```

预期输出应显示两个容器都在运行状态：
- nofx-dev-frontend-display
- nofx-dev-backend-display

## 访问服务

- **前端界面**: http://localhost:3300
- **后端API**: http://localhost:8888
- **后端Swagger UI**（如果可用）: http://localhost:8888/swagger

## 特殊功能

### 浏览器弹窗功能

后端容器配置了 `GUARDIAN_DISPLAY_ENABLED=true`，这意味着Chrome将以非无头模式运行，会弹出浏览器窗口，便于测试浏览器自动化功能。

### 数据库访问

- 数据库文件: `E:\AI\nofx_dev\data\data.db`
- 该数据库为可读写模式
- 通过容器卷映射: `./data:/app/data`

## 管理命令

### 启动服务
```bash
docker-compose -f docker-compose.dev.display.yml up -d
```

### 停止服务
```bash
docker-compose -f docker-compose.dev.display.yml down
```

### 查看日志
```bash
docker-compose -f docker-compose.dev.display.yml logs -f
```

### 重新构建并启动
```bash
docker-compose -f docker-compose.dev.display.yml up --build -d
```

## 故障排除

### 容器启动失败

1. 检查 Docker 是否正在运行
2. 检查端口 3300 和 8888 是否已被占用
3. 检查所需镜像是否存在

### 前端无法访问

1. 确认前端容器正在运行
2. 检查防火墙设置
3. 确认端口映射正确

### 后端服务异常

1. 查看后端容器日志
2. 检查数据库连接
3. 确认环境变量设置正确

### 浏览器弹窗问题

1. 确认后端容器设置了 `GUARDIAN_DISPLAY_ENABLED=true`
2. 检查系统是否安装了Chrome/Chromium
3. 确认Docker有权限显示GUI（Linux需要X11转发）

## 安全注意事项

- 此配置仅供开发和测试使用
- Chrome非无头模式可能暴露敏感信息
- 生产环境请勿使用此配置
- 数据库文件位于本地，注意备份

## 清理

如果需要彻底清理 nofx-dev-web 分组：

```bash
docker-compose -f docker-compose.dev.display.yml down -v --remove-orphans
```

## 维护

### 更新镜像

```bash
# 拉取最新镜像
docker pull setup-nofx-dev-frontend:latest
docker pull nofx_dev-backend-display:latest

# 重新启动服务
docker-compose -f docker-compose.dev.display.yml up -d
```

### 备份数据库

定期备份 `E:\AI\nofx_dev\data\data.db` 文件以防数据丢失。

---

**注意**: 请确保不要同时运行多个nofx相关容器组以避免数据库冲突。nofx-dev-web分组中的数据库为可读写模式，使用时请注意