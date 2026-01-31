# 完整系统部署指南

按照本指南，您将部署一个完整的系统，其中：
- 前端运行在宿主机的3300端口
- 后端运行在宿主机的8888端口  
- 币安数据获取插件运行在Docker容器中
- 所有与币安API的通信都将通过币安插件管道进行

## 部署步骤

### 1. 启动币安代理服务（Docker容器内）

在项目根目录运行以下命令来启动币安代理服务：

```bash
docker-compose -f docker-compose.binance-proxy.yml up --build -d
```

这将：
- 构建币安代理服务镜像
- 启动币安代理服务容器
- 将容器的8080端口映射到宿主机的8081端口

### 2. 配置环境变量

确保 `.env` 文件中包含以下配置：

```bash
# 启用币安代理模式
USE_BINANCE_PROXY=true

# 代理服务地址（由于币安代理运行在Docker中，其他服务运行在宿主机上）
BINANCE_PROXY_URL=http://localhost:8081

# 后端API服务器端口
NOFX_BACKEND_PORT=8888

# 前端Web界面端口
NOFX_FRONTEND_PORT=3300
```

### 3. 启动后端服务（宿主机）

在项目根目录运行：

```bash
go run main.go
```

后端服务将在8888端口启动，并通过币安代理服务与币安API通信。

### 4. 启动前端服务（宿主机）

在项目根目录运行：

```bash
cd web
npm install
npm run dev
```

前端服务将在3300端口启动，通过后端服务与币安代理服务通信，最终与币安API交换数据。

## 系统架构

```
前端 (localhost:3300) 
    ↓ (HTTP请求)
后端 (localhost:8888)
    ↓ (通过环境变量BINANCE_PROXY_URL控制)
币安代理服务 (localhost:8081/Docker容器内)
    ↓ (实际的币安API调用)
币安API
```

## 验证部署

1. **验证币安代理服务**：
   ```bash
   curl http://localhost:8081/health
   ```

2. **验证后端服务**：
   ```bash
   curl http://localhost:8888/api/health
   ```

3. **访问前端**：
   打开浏览器访问 http://localhost:3300

## 故障排除

### 1. 检查Docker容器是否运行
```bash
docker ps
```

你应该看到名为 `binance-proxy` 的容器正在运行。

### 2. 检查币安代理服务日志
```bash
docker logs binance-proxy
```

### 3. 检查端口占用
确保3300、8888和8081端口未被其他服务占用。

### 4. 网络连接测试
确保宿主机上的后端服务能够访问Docker容器中的币安代理服务。

## 重新配置

如果需要禁用代理模式（直接连接币安API），只需修改 `.env` 文件：

```bash
USE_BINANCE_PROXY=false
```

然后重启后端服务。

## 停止服务

停止币安代理服务：
```bash
docker-compose -f docker-compose.binance-proxy.yml down
```

停止后端服务：按 Ctrl+C

停止前端服务：按 Ctrl+C