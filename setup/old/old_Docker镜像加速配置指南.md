# Docker镜像加速配置指南

本指南将帮助您配置Docker镜像加速器，以解决在中国大陆地区拉取Docker镜像时遇到的网络问题。

## 配置Docker Desktop

### Windows系统配置步骤：

1. 打开Docker Desktop
2. 点击右上角的设置图标（Settings）
3. 选择 "Docker Engine" 选项
4. 在右侧的JSON配置中，替换原有的配置为以下内容：

```json
{
 "registry-mirrors": [
   "https://hub-mirror.c.163.com",
   "https://mirror.baidubce.com",
   "https://docker.mirrors.ustc.edu.cn",
   "https://registry.docker-cn.com"
 ]
}
```

5. 点击 "Apply & Restart" 重启Docker服务

## 配置Docker Daemon（Linux系统）

如果您在Linux系统上使用Docker，请按以下步骤配置：

1. 创建或编辑 `/etc/docker/daemon.json` 文件：

```bash
sudo nano /etc/docker/daemon.json
```

2. 添加以下内容：

```json
{
 "registry-mirrors": [
   "https://hub-mirror.c.163.com",
   "https://mirror.baidubce.com",
   "https://docker.mirrors.ustc.edu.cn",
   "https://registry.docker-cn.com"
 ]
}
```

3. 重启Docker服务：

```bash
sudo systemctl restart docker
```

## 验证配置

配置完成后，您可以运行以下命令验证配置是否生效：

```bash
docker info
```

在输出中查找 "Registry Mirrors" 部分，确认加速器地址已列出。

## 项目中的Dockerfile优化

我们已经在项目的Dockerfiles中添加了Alpine Linux的国内镜像源，以加快包的安装速度：

- `docker/Dockerfile.backend`
- `docker/Dockerfile.backend.dev` 
- `docker/Dockerfile.guardian`

这些配置将使用阿里云的Alpine镜像源，提高构建速度。

## 故障排除

如果仍然遇到网络问题，请尝试：

1. 检查防火墙设置
2. 确认网络连接正常
3. 尝试更换不同的镜像加速器地址
4. 清理Docker构建缓存：

```bash
docker builder prune
```

## 重新构建项目

配置完成后，您可以重新尝试构建项目：

```bash
# 开发环境
docker-compose -f docker-compose.dev.with-guardian.yml up -d --build

# 生产环境
docker-compose -f docker-compose.with-guardian.yml up -d --build
```

使用 `--build` 参数确保重新构建镜像。