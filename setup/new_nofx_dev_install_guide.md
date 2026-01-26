# nofx-dev 安装与部署指南

## 概述

本指南介绍如何安装和部署 nofx-dev（开发版），该版本适用于日常开发和调试。

## 分组组成

nofx-dev 分组包含以下服务：

- **前端**: nofx-dev-frontend (默认端口 3300)
- **后端**: nofx-dev-backend (默认端口 8888)
- **功能**: 标准开发模式，支持热更新
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
git clone <dev-repository-url>
cd nofx_dev
```

### 2. 确保已有镜像

确保以下镜像已在本地存在：

- `setup-nofx-dev-frontend:latest`
- `nofx_dev-backend:latest`

如果不存在，请先构建这些镜像。

### 3. 启动 nofx-dev 分组

```bash
# 在项目根目录下运行
docker-compose -f docker-compose.dev.yml up -d
```

### 4. 验证部署

检查容器是否正在运行：

```bash
docker-compose -f docker-compose.dev.yml ps
```

预期输出应显示两个容器都在运行状态：
- nofx-dev-frontend
- nofx-dev-backend

## 访问服务

- **前端界面**: http://localhost:3300
- **后端API**: http://localhost:8888
- **后端Swagger UI**（如果可用）: http://localhost:8888/swagger

## 特殊功能

### 热更新

- 前端支持实时热更新
- 代码修改后自动重新加载
- 适合快速开发迭代

### 开发调试

- 详细的日志输出
- 错误信息更丰富
- 便于调试和问题排查

### 数据库访问

- 数据库文件: `E:\AI\nofx_dev\data\data.db`
- 该数据库为可读写模式
- 通过容器卷映射: `./data:/app/data`

## 管理命令

### 启动服务
```bash
docker-compose -f docker-compose.dev.yml up -d
```

### 停止服务
```bash
docker-compose -f docker-compose.dev.yml down
```

### 查看日志
```bash
docker-compose -f docker-compose.dev.yml logs -f
```

### 重新构建并启动
```bash
docker-compose -f docker-compose.dev.yml up --build -d
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

## 安全注意事项

- 此配置仅供开发和调试使用
- 包含详细日志，可能泄露敏感信息
- 生产环境请勿使用此配置

## 清理

如果需要彻底清理 nofx-dev 分组：

```bash
docker-compose -f docker-compose.dev.yml down -v --remove-orphans
```

## 维护

### 更新镜像

```bash
# 拉取最新镜像
docker pull setup-nofx-dev-frontend:latest
docker pull nofx_dev-backend:latest

# 重新启动服务
docker-compose -f docker-compose.dev.yml up -d
```

### 备份数据库

定期备份 `E:\AI\nofx_dev\data\data.db` 文件以防数据丢失。

---

**注意**: 请确保不要同时运行多个nofx相关容器组以避免数据库冲突。nofx-dev分组中的数据库为可读写模式，使用时请注意