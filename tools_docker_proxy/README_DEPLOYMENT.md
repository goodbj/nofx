# 透明代理 Docker 部署指南

## 环境要求

- Docker Desktop 已安装并运行
- Docker Compose 已安装

## 端口配置

- 透明代理服务默认端口：8081

## 部署步骤

### 1. 环境准备

确保Docker服务正在运行：

```bash
docker --version
docker-compose --version
```

### 2. 构建并启动服务（支持热更新）

```bash
# 进入项目目录
cd tools_docker_proxy

# 使用热更新配置部署
docker-compose -f docker-compose.hot.yml up -d --build
```

### 3. 验证部署

```bash
# 查看容器状态
docker ps --filter "name=tools_docker_proxy"

# 查看服务日志
docker logs tools_docker_proxy --tail 20

# 测试健康检查
curl http://localhost:8081/health
```

## 配置文件说明

### Dockerfile.hot
- 专为热更新设计的Docker镜像
- 使用fswatch监控文件变化
- 自动重新编译和重启服务

### docker-compose.hot.yml
- 定义热更新服务配置
- 将本地代码目录挂载到容器中
- 映射端口8081到容器内部

## 管理命令

```bash
# 查看实时日志
docker logs -f tools_docker_proxy

# 重启服务
docker restart tools_docker_proxy

# 停止服务
docker-compose -f docker-compose.hot.yml down

# 清理并重建
docker-compose -f docker-compose.hot.yml down
docker rmi nofx-proxy:latest
docker-compose -f docker-compose.hot.yml up -d --build
```

## 故障排除

### 端口冲突
如果端口8081已被占用，可以在`.env`文件中修改：
```
PROXY_PORT=8082
```

### 构建失败
检查Go代码是否有语法错误：
```bash
go build .
```

### 容器无法启动
查看详细日志：
```bash
docker logs tools_docker_proxy
```

## 热更新机制

当您修改Go源代码文件时，容器会：
1. 自动检测文件变化
2. 重新编译应用程序
3. 重启服务
4. 保持服务连续性

## 透明代理功能

此代理服务实现以下功能：
- 纯透传原则，不处理业务逻辑
- 验证与远程服务的连接性
- 传输认证信息
- 将请求转发给远程服务
- 将响应返回给nofx原生方法进行处理