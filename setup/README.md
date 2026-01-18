# NOFX Docker 部署安装文档

## 概述

本文档提供了完整的NOFX项目Docker部署指南，包含常见问题解决方案和自动化脚本，帮助您快速部署开发版和稳定版环境。

## 文件结构说明

本目录(`setup`)包含以下部署相关文件：

- `docker-compose.stable.yml` - 稳定版部署配置文件
- `docker-compose.dev.watch.yml` - 开发版部署配置文件
- `deploy-stable.bat` - 稳定版一键部署脚本
- `deploy-dev.bat` - 开发版一键部署脚本
- `system-status.bat` - 双系统状态检查脚本

本文档涵盖了以下部署场景：
- 开发版部署（支持热更新，适合开发调试）
- 稳定版部署（生产环境，使用预构建镜像）

**重要说明：** 两个版本部署在不同的目录中：
- **稳定版**：`E:\AI\nofx` （前端端口3000，后端端口8080）
- **开发版**：`E:\AI\nofx_Dev` （前端端口3300，后端端口8888）

## 系统要求

- Docker Desktop (Windows) 或 Docker Engine (Linux/Mac)
- Docker Compose v2.x 或更高版本
- 至少4GB可用内存
- 10GB可用磁盘空间

## 预配置步骤

### 1. 项目目录结构

**稳定版目录结构** (E:\AI\nofx)：
```
E:\AI\nofx
├── data/ (数据库文件目录)
│   └── data.db
├── docker/
├── docker-compose.stable.yml
└── .env
```

**开发版目录结构** (E:\AI\nofx_Dev)：
```
E:\AI\nofx_Dev
├── data/ (数据库文件目录)
│   └── data.db
├── web/ (前端源码)
├── docker/
├── docker-compose.dev.watch.yml
└── .env
```

**注意**：两个版本在不同目录中完全独立，可以根据需要安装其中一个或两个都安装。

### 2. 环境变量配置 (.env)

#### 重要说明
- `.env.example` 是官方提供的 `.env` 标准示例，在程序调试过程中不允许修改
- 在重新部署服务时，`.env` 中的 `RSA_PRIVATE_KEY=` 要严格按照下面格式修改

#### RSA密钥格式要求
**必须注意！！！下面是标准的密钥格式，带\n 换行符，要在同一行，一定要严格按照这个格式添加：**
```
# 示例 RSA_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\nYOUR_KEY_HERE\n-----END RSA PRIVATE KEY-----
```

创建 `.env` 文件并填入以下内容：

```env
# NOFX 环境配置文件
TZ=Asia/Shanghai
NOFX_BACKEND_PORT=8888
NOFX_FRONTEND_PORT=3300

# 安全配置 - 请替换为强密码
JWT_SECRET=your_strong_jwt_secret_here_replace_with_random_string_at_least_32_chars
DATA_ENCRYPTION_KEY=your_data_encryption_key_here_replace_with_32_char_hex_string
RSA_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA7ZAy8I/xU3/pv7MZ25nHgRD1XSrCMHtH2OL/qwqCvkvGCRhS\nXoVSsimeeBB/PNSU3iYD6BrXaYk7YKnuzHhftiRq/D4I+itjY0jf24UyJ6KmfT6s\nKQAjWxTsOEQUktrfFXaGIg2e+KhpYXnZze2PUDnNWxNqfi6PEEqlOEbd/fzzThY0\nem6T9RH7wawScVUR6w3v3olO+BUsu02kHXDkQ1YPPcrHgoTQ8pxHUm/3N/feDqwn\n/Ub4kvbg88mZvEgRldRgAECtXBMk1CKzXQNPRYxTSckKffSUVqCEzrPkix/TFVGR\nQE7iaCAEBvT5aqDo+dMsxkDULnsQGC/4WXcPSwIDAQABAoIBADXVtS19uSb+eDao\nfCYfMa5GbQwJags5jL0SJ/UXQyyjmEOsXtIrrWNReidkOalL1VaIT99T4df5MNsF\nd2efqbTpiNMTrc4fcfzoYU5qX0TLH6aHQtVhwiFcWvGfP/hNoDtJajkiVBGufH8J\n8Xkwqgb4qlhGzJ2+qE39VHat3JW3QpXcuXH/PZoe1IJq4otwDujUdZGYRYSH/ZlF\nbhyIQfL0RkFknySK3XgtrloEyawPPbPs92ofxNH0mAOvuRYMFExbp00r0fP7QL/Y\nKyAe5DsWuYOlzlKRusrN0Y3sJFhIyoRr25w0p8Qez2W1km4ryj31v+qarCrlj2le\n7Dtio0kCgYEA82VZRhdKR5Jti5TOg/WmNL58OUniN1BrfIIHRSOGTFkqI7XkzQXV\nETzZoJcXVq8+HBiaKVkTtc4DHxQS9RwXzqxyGkSSiBovixbPKhPc+D/m0FZTaqnP\nkXGT9QVdY00PapZb3P3gFK3VlXvbh/dDgPFFPFpaRifaIQHvr0VDPGMCgYEA+d2H\nRlY8iEddjly1nk0qmJa4Bl8MIrdQbmpvakI6q3DizeoMvGduyjK6ciDy++T9Zqb7\niYeOqaNFniuqZCmBukOPzVD2QXYHROm9W08P1tQXCxfyAmKcLxLJUA5TrpPt792T\n2yDpCK7lRATAlRfdip2j4DHwR9NjBVMBw5ISUfkCgYBW/1XOkMqTFIqlRpYeYrJ6\nzc9XJsp93PfedBenJdB9/6zpQL28bqY+2BItrXPBHzhDEKQhvV4nMLC67hDsnZMA\n43CRZQs/LKTrwUZhEuJ7tVOKCiEc0f+ITCGHhdhggw3MmlvRfMkYex4JpVDNo5r0\nPsjxjpYP13THMYr7ifVDYwKBgHWyw2D/iD4Nl+VSiH7MDK+Z94+QwC+uOCX63wan\nselGIKAsitlIw6hdYvQVzz+Wq0Lqj3xGLY59CXMrUHUkFCbAYoGtjIJjbaMpk3fq\ncySX/U7NdcNn3fhSmh+q0AJhTmh58Ib9JqhfckGrF2hjuIjuHt6hx3Sd/3vnkOIl\n8ZlJAoGAJISsNmC58IXZat/BeGGtoVKxCB++F9R5MBszmJdwuMrL+DYeWqpXLC54\naIJaQrYj7KjQ/wXP9PM3pndll8RvsbM7kOwJ56FnJk+1wZxGosXENhgHLbkZoBBD\nHHpT8ePjQ87geH1/X8wbI7YIvAg/Rh8bDMkfDBA/kmIC06zUvFI=\n-----END RSA PRIVATE KEY-----\n"
```

## 开发版部署 (E:\AI\nofx_Dev)

### 部署前准备
1. 确保在 `E:\AI\nofx_Dev` 目录中
2. 确保 `data` 目录和 `data.db` 文件存在
3. 确保 `.env` 文件配置正确

### 自动化部署 (推荐)

运行一键部署脚本：

```bash
# Windows
cd E:\AI\nofx_Dev
.\setup\install-dev.bat
```

### 手动部署步骤

1. 构建后端镜像：
```bash
cd E:\AI\nofx_Dev
docker-compose -f docker-compose.dev.watch.yml build nofx-trading-dev-watch
```

2. 构建前端镜像：
```bash
docker-compose -f docker-compose.dev.watch.yml build nofx-frontend-dev-watch
```

3. 启动服务：
```bash
docker-compose -f docker-compose.dev.watch.yml up -d
```

4. 验证部署：
```bash
# 检查服务状态
docker-compose -f docker-compose.dev.watch.yml ps

# 查看日志
docker logs nofx-trading-dev-watch
docker logs nofx-frontend-dev-watch
```

## 稳定版部署 (E:\AI\nofx)

### 部署前准备
1. 确保在 `E:\AI\nofx` 目录中
2. 确保 `data` 目录和 `data.db` 文件存在
3. 确保 `.env` 文件配置正确

### 自动化部署 (推荐)

运行一键部署脚本：

```bash
# Windows
cd E:\AI\nofx
.\setup\install-prod.bat
```

### 手动部署步骤

1. 构建稳定版镜像：
```bash
cd E:\AI\nofx
docker-compose -f docker-compose.stable.yml build
```

2. 启动稳定版服务：
```bash
docker-compose -f docker-compose.stable.yml up -d
```

## 常见问题及解决方案

### 1. RSA_PRIVATE_KEY 格式错误

**症状**：
- 项目无法启动
- 报错与加密相关的错误

**解决方案**：
确保 `.env` 文件中的 `RSA_PRIVATE_KEY` 使用正确格式，包含 `\n` 换行符：

```env
RSA_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\nMIIEogIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----\n"
```

### 2. exec format error

**症状**：
```
exec ./nofx: exec format error
```

**原因**：
- Go交叉编译参数设置错误
- Windows/Linux二进制格式不匹配

**解决方案**：
在Dockerfile.backend中确保设置正确的环境变量：
```dockerfile
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
```

### 3. 前端 Rollup 依赖问题

**症状**：
```
Cannot find module @rollup/rollup-linux-x64-musl
Cannot find module @rollup/rollup-linux-x64-gnu
```

**解决方案**：
- 使用 `ROLLUP_NATIVE=0` 环境变量绕过原生二进制依赖
- 或使用正确的平台特定模块

### 4. 前后端通信失败

**症状**：
- 前端无法连接后端API
- 网络代理错误

**解决方案**：
- 确保docker-compose中设置正确的 `API_HOST` 环境变量
- 验证容器网络配置

### 5. 端口被占用

**症状**：
- 无法绑定到指定端口

**解决方案**：
```bash
# 检查端口占用
netstat -ano | findstr :8888
netstat -ano | findstr :3300

# 结束占用进程
taskkill /PID <PID> /F
```

## 故障排除

### 检查服务状态

对于稳定版 (在 E:\AI\nofx 目录下)：
```bash
docker-compose -f docker-compose.stable.yml ps
docker-compose -f docker-compose.stable.yml logs nofx-trading
docker-compose -f docker-compose.stable.yml logs nofx-frontend
```

对于开发版 (在 E:\AI\nofx_Dev 目录下)：
```bash
docker-compose -f docker-compose.dev.watch.yml ps
docker-compose -f docker-compose.dev.watch.yml logs nofx-trading-dev-watch
docker-compose -f docker-compose.dev.watch.yml logs nofx-frontend-dev-watch
```

### 重启服务

重启稳定版 (在 E:\AI\nofx 目录下)：
```bash
docker-compose -f docker-compose.stable.yml down
docker-compose -f docker-compose.stable.yml up -d
```

重启开发版 (在 E:\AI\nofx_Dev 目录下)：
```bash
docker-compose -f docker-compose.dev.watch.yml down
docker-compose -f docker-compose.dev.watch.yml up -d
```

### 清理Docker资源
```bash
# 清理未使用的容器、网络、镜像
docker system prune -f

# 清理构建缓存
docker builder prune -f
```

## 访问服务

### 稳定版 (E:\AI\nofx)
- **前端界面**：http://localhost:3000
- **后端API**：http://localhost:8080
- **后端健康检查**：http://localhost:8080/api/health
- **容器名称**：nofx-trading, nofx-frontend

### 开发版 (E:\AI\nofx_Dev)
- **前端界面**：http://localhost:3300
- **后端API**：http://localhost:8888
- **后端健康检查**：http://localhost:8888/api/health
- **容器名称**：nofx-trading-dev-watch, nofx-frontend-dev-watch

### 两套系统独立运行说明
当同时运行开发版和稳定版时：
- **端口分离**：
  - 稳定版使用 8080(后端)/3000(前端) 端口
  - 开发版使用 8888(后端)/3300(前端) 端口
- **容器名称独立**：
  - 使用不同的容器名称，避免冲突
- **数据持久化**：
  - 稳定版数据库：E:\AI\nofx\data\data.db
  - 开发版数据库：E:\AI\nofx_Dev\data\data.db
- **功能特性**：
  - 开发版配置了卷挂载，支持代码热更新
  - 两套系统完全独立运行，互不影响
- **使用建议**：
  - 日常使用：访问稳定版 http://localhost:3000
  - 开发测试：访问开发版 http://localhost:3300（支持热更新）

## 独立部署说明

您可以根据需要单独部署稳定版或开发版，两个系统完全独立，可以单独安装和使用。

### 何时选择稳定版
- 用于日常生产环境
- 需要稳定的交易环境
- 使用预构建的稳定镜像
- 访问地址：http://localhost:3000

### 何时选择开发版
- 用于开发和测试新功能
- 需要代码热更新功能
- 进行调试和实验
- 访问地址：http://localhost:3300

### 同时运行两个版本
如果需要同时运行两个版本：
- 稳定版使用端口 3000(前端) 和 8080(后端)
- 开发版使用端口 3300(前端) 和 8888(后端)
- 数据库文件分别存储在各自目录中，互不干扰

## 验证系统运行状态

### 检查稳定版 (E:\AI\nofx)
在 E:\AI\nofx 目录下运行：
```bash
docker-compose -f docker-compose.stable.yml ps
```

### 检查开发版 (E:\AI\nofx_Dev)
在 E:\AI\nofx_Dev 目录下运行：
```bash
docker-compose -f docker-compose.dev.watch.yml ps
```

### 访问服务
- 稳定版前端：http://localhost:3000
- 稳定版后端：http://localhost:8080
- 开发版前端：http://localhost:3300
- 开发版后端：http://localhost:8888

### 重要原则
- 在开发版开发的时候不要修改稳定版的任何内容，包括 docker中的实例
- `.env.example` 是官方给的 `.env` 标准示例，以后在程序调试过程中也不允许修改

## 卸载/重置

### 卸载稳定版 (在 E:\AI\nofx 目录下)
```bash
# 停止并删除容器
docker-compose -f docker-compose.stable.yml down

# 删除相关镜像（可选）
docker rmi ghcr.io/nofxaios/nofx/nofx-backend:stable ghcr.io/nofxaios/nofx/nofx-frontend:stable
```

### 卸载开发版 (在 E:\AI\nofx_Dev 目录下)
```bash
# 停止并删除容器
docker-compose -f docker-compose.dev.watch.yml down

# 删除相关镜像（可选）
docker rmi nofx_dev-nofx-frontend-dev-watch nofx_dev-nofx-trading-dev-watch
```