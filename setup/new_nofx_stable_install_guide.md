# nofx-stable 安装与部署指南

## 概述

本指南介绍如何安装和部署 nofx-stable（稳定版），该版本适用于生产环境或稳定的开发测试。

## 分组组成

nofx-stable 分组包含以下服务：

- **前端**: nofx-stable-frontend (默认端口 3000)
- **后端**: nofx-stable-backend (默认端口 8080)
- **功能**: 标准运行模式，无浏览器弹窗
- **文件目录**: E:\AI\nofx
- **数据库目录**: E:\AI\nofx\data\data.db（可读写）

## 前提条件

- Docker Desktop 已安装并运行
- Docker Compose 已安装
- 至少 4GB 可用磁盘空间
- 稳定的互联网连接

## 安装步骤

### 1. 克隆或下载项目

```bash
git clone <stable-repository-url>
cd nofx
```

### 2. 确保已有镜像

确保以下镜像已在本地存在：

- `setup-nofx-stable-frontend:latest`
- `setup-nofx-stable-backend:latest`

如果不存在，请先构建这些镜像。

### 3. 启动 nofx-stable 分组

```bash
# 在项目根目录下运行
docker-compose -f docker-compose.stable.yml up -d
```

### 4. 验证部署

检查容器是否正在运行：

```bash
docker-compose -f docker-compose.stable.yml ps
```

预期输出应显示两个容器都在运行状态：
- nofx-stable-frontend
- nofx-stable-backend

## 访问服务

- **前端界面**: http://localhost:3000
- **后端API**: http://localhost:8080
- **后端Swagger UI**（如果可用）: http://localhost:8080/swagger

## 特殊功能

### 稳定运行

- 标准运行模式，性能优化
- 无浏览器弹窗干扰
- 适合长期稳定运行

### 数据库访问

- 数据库文件: `E:\AI\nofx\data\data.db`
- 该数据库为可读写模式
- 通过容器卷映射: `./data:/app/data`

## 管理命令

### 启动服务
```bash
docker-compose -f docker-compose.stable.yml up -d
```

### 停止服务
```bash
docker-compose -f docker-compose.stable.yml down
```

### 查看日志
```bash
docker-compose -f docker-compose.stable.yml logs -f
```

### 重新构建并启动
```bash
docker-compose -f docker-compose.stable.yml up --build -d
```

## 故障排除

### 容器启动失败

1. 检查 Docker 是否正在运行
2. 检查端口 3000 和 8080 是否已被占用
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

- 此配置适用于稳定运行环境
- 无浏览器弹窗，减少安全风险
- 生产环境推荐使用此配置

## 清理

如果需要彻底清理 nofx-stable 分组：

```bash
docker-compose -f docker-compose.stable.yml down -v --remove-orphans
```

## 维护

### 更新镜像

```bash
# 拉取最新镜像
docker pull setup-nofx-stable-frontend:latest
docker pull setup-nofx-stable-backend:latest

# 重新启动服务
docker-compose -f docker-compose.stable.yml up -d
```

### 备份数据库

定期备份 `E:\AI\nofx\data\data.db` 文件以防数据丢失。

---

**注意**: 请确保不要同时运行多个nofx相关容器组以避免数据库冲突。nofx-stable分组中的数据库为可读写模式，使用时请注意