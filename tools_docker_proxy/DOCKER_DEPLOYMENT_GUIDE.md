# 🐳 透明代理服务 Docker 部署指南

## 📋 环境要求
- Docker Desktop for Windows
- Windows 10/11 
- 至少 4GB 可用内存

## 🚀 部署步骤

### 1. 启动 Docker
首先确保 Docker Desktop 已安装并运行：
- 在系统托盘找到 Docker 图标
- 确保状态显示为 "Running"
- 或通过命令 `docker info` 验证

### 2. 部署选项

#### 🏭 生产环境部署 (推荐)
```
# 方法1: 使用部署脚本
双击运行 deploy_docker.bat

# 方法2: 使用 Docker Compose
docker-compose up -d --build
```

#### 🔧 开发环境部署 (热更新)
```
# 开发模式部署（支持代码实时更新）
docker-compose -f docker-compose.dev.yml up -d --build
```

### 3. Docker 容器信息

容器参数说明：
- **容器名称**: `tools_docker_proxy`
- **端口映射**: `8081:8081` (主机端口:容器端口)
- **网络**: `nofx-proxy-network` (桥接网络)
- **重启策略**: `unless-stopped` (异常停止时自动重启)

### 4. 管理命令

#### 常用管理脚本
```
# 启动管理菜单
双击运行 docker_manager.bat
```

#### 手动管理命令
```bash
# 查看容器状态
docker ps -a --filter "name=tools_docker_proxy"

# 查看实时日志
docker logs -f tools_docker_proxy

# 停止容器
docker stop tools_docker_proxy

# 重启容器
docker restart tools_docker_proxy

# 删除容器
docker rm tools_docker_proxy

# 清理所有相关资源
docker stop tools_docker_proxy
docker rm tools_docker_proxy
docker rmi nofx-proxy:latest
docker network rm nofx-proxy-network
```

### 5. 健康检查

部署完成后，可以通过以下方式验证服务状态：

```bash
# 检查端口占用
netstat -an | findstr :8081

# HTTP 健康检查
curl http://localhost:8081/health
# 应返回: OK

# Docker 日志检查
docker logs tools_docker_proxy --tail 10
```

### 6. 性能优化特性

部署的透明代理服务包含以下优化：

✅ **连接池优化**
- 最大空闲连接: 200
- 每主机最大空闲连接: 20
- 连接超时: 120秒

✅ **并发控制**
- 最大并发请求数: 100
- 防止资源耗尽

✅ **异步日志处理**
- 非阻塞日志写入
- 日志队列缓冲

✅ **内存池优化**
- 32KB 缓冲区复用
- 减少内存分配开销

✅ **响应压缩**
- 启用 HTTP 压缩
- 减少网络传输量

### 7. 故障排除

#### 常见问题

**问题1: 端口被占用**
```bash
# 查找占用8081端口的进程
netstat -ano | findstr :8081
# 杀死占用进程
taskkill /f /pid [进程ID]
```

**问题2: Docker 无法启动**
- 确保 Docker Desktop 已安装
- 检查 Windows 虚拟化功能是否启用
- 重启 Docker Desktop 服务

**问题3: 容器启动失败**
```bash
# 查看详细错误信息
docker logs tools_docker_proxy
# 检查镜像构建日志
docker build -t nofx-proxy:latest .
```

**问题4: 网络连接问题**
```bash
# 检查 Docker 网络
docker network ls
# 重新创建网络
docker network create nofx-proxy-network
```

### 8. 监控和维护

#### 日志监控
```bash
# 实时查看日志
docker logs -f tools_docker_proxy

# 查看最近日志
docker logs tools_docker_proxy --tail 50

# 查看带时间戳的日志
docker logs -t tools_docker_proxy
```

#### 资源监控
```bash
# 查看容器资源使用
docker stats tools_docker_proxy

# 查看容器详细信息
docker inspect tools_docker_proxy
```

### 9. 安全注意事项

- 容器运行在隔离环境中
- 端口映射仅暴露必要的 8081 端口
- 使用非 root 用户运行（如需）
- 定期更新基础镜像

### 10. 备份和恢复

```bash
# 备份容器配置
docker inspect tools_docker_proxy > proxy_config_backup.json

# 导出镜像
docker save nofx-proxy:latest > proxy_image_backup.tar

# 导入镜像
docker load < proxy_image_backup.tar
```

---
**部署完成标志**: 能够通过 `http://localhost:8081/health` 访问并返回 "OK"