# 安装Docker并部署透明代理服务

## 第一步：安装Docker Desktop

由于您的系统尚未安装Docker，您需要首先安装Docker Desktop：

1. 访问 [Docker官网](https://www.docker.com/products/docker-desktop/)
2. 下载适用于Windows的Docker Desktop安装包
3. 运行安装程序并按照提示完成安装
4. 启动Docker Desktop应用程序

## 第二步：验证Docker安装

安装完成后，在PowerShell或命令提示符中运行以下命令验证安装：

```powershell
docker --version
docker-compose --version
```

## 第三步：部署透明代理服务

一旦Docker安装并运行，您可以使用以下任一方式部署透明代理服务：

### 方法1：使用快速启动脚本
```powershell
cd e:\AI\nofx_dev\tools_docker_proxy
.\QUICK_START.bat
```

### 方法2：使用协调启动脚本
```powershell
cd e:\AI\nofx_dev\tools_docker_proxy
.\START_WITH_MAIN_PROJECT.bat
```

### 方法3：手动部署
```powershell
cd e:\AI\nofx_dev\tools_docker_proxy
docker-compose -f docker-compose.hot.yml down
docker-compose -f docker-compose.hot.yml up -d --build
```

## 第四步：验证部署

部署完成后，验证服务是否正常运行：

```powershell
# 检查容器状态
docker ps --filter "name=tools_docker_proxy"

# 检查服务健康状态
curl http://localhost:8081/health

# 查看服务日志
docker logs tools_docker_proxy --tail 20
```

## 透明代理服务配置详情

- **容器名称**: tools_docker_proxy
- **端口**: 8081（与主项目.env文件中的配置一致）
- **功能**: 透明代理服务，用于绕过地区限制访问受限制的交易所API
- **热更新**: 已启用，修改Go源代码时服务会自动重新加载
- **环境变量**: 
  - PROXY_PORT=8081
  - PORT=8081

## 与主项目集成

确保主项目中的以下配置与透明代理服务匹配：

```env
USE_BINANCE_PROXY=true
BINANCE_PROXY_URL=http://localhost:8081
BINANCE_PROXY_PORT=8081
```

这些配置在 `e:\AI\nofx_dev\.env` 文件中。

## 故障排除

### 如果遇到端口冲突
- 检查端口8081是否被其他服务占用
- 可以使用 `netstat -ano | findstr :8081` 命令查看端口占用情况

### 如果容器无法启动
- 查看详细日志：`docker logs tools_docker_proxy`
- 检查Docker资源分配是否足够

### 如果无法访问服务
- 确认防火墙未阻止8081端口
- 检查Docker服务是否正在运行

## 管理命令

- **停止服务**: `docker-compose -f docker-compose.hot.yml down`
- **查看日志**: `docker logs -f tools_docker_proxy`
- **重启服务**: `docker restart tools_docker_proxy`
- **清理并重建**: 
  ```powershell
  docker-compose -f docker-compose.hot.yml down
  docker rmi nofx-proxy:latest
  docker-compose -f docker-compose.hot.yml up -d --build
  ```

## 热更新机制

透明代理服务支持热更新，当您修改Go源代码文件时：
1. 容器会自动检测文件变化
2. 重新编译应用程序
3. 重启服务
4. 保持服务连续性

这使得开发和调试变得更加高效。