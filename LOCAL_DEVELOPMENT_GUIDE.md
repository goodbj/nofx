# NOFX 本地开发指南

本文档介绍了如何在本地环境中运行 NOFX 项目进行实时开发和测试，无需每次都构建 Docker 镜像。

## 系统要求

- Go 1.19+
- Node.js 18+
- npm (随 Node.js 一起安装)

## 快速开始

### Windows 用户

运行批处理脚本：

```cmd
local_dev_setup.bat
```

这将：
1. 检查 Go、Node.js 和 npm 是否已安装
2. 创建 `.env` 配置文件（如不存在）
3. 创建 `data` 目录用于存储数据库
4. 安装 Go 依赖
5. 安装前端依赖
6. 同时启动后端和前端服务器

### macOS/Linux 用户

运行以下命令：

```bash
# 安装 Go 依赖
go mod tidy

# 启动后端服务（在第一个终端窗口）
go run main.go

# 在另一个终端窗口中启动前端服务
cd web
npm install
npm run dev
```

## 开发工作流

### 后端开发

1. 修改 Go 代码（例如在 `api/`, `trader/`, `decision/`, `backtest/` 等目录）
2. 在后端终端窗口按 `Ctrl+C` 停止当前进程
3. 重新运行 `go run main.go` 以应用更改
4. 测试新功能

### 前端开发

1. 修改前端代码（位于 `web/src/` 目录）
2. 保存文件后，Vite 会自动热重载更改
3. 在浏览器中即时查看更改

## 服务地址

- **前端界面**: http://localhost:3300
- **后端 API**: http://localhost:8888
- **API 文档**: http://localhost:8888/swagger/index.html

## 环境配置

### 复制环境变量文件

```bash
cp .env.example .env
```

### 数据库

- 默认使用 SQLite，数据文件存储在 `data/data.db`
- 如果需要其他数据库，请修改 `.env` 文件中的数据库配置

## 测试您的更改

由于您最近添加了动态止损、止盈和部分平仓功能，您可以：

1. 在后端代码中修改这些功能的实现
2. 保存文件并重启后端服务
3. 通过 API 调用或前端界面测试新功能
4. 观察日志以验证功能是否按预期工作

## 故障排除

### 端口被占用

如果端口 8888 或 3300 已被占用，请修改 `.env` 文件中的 `NOFX_BACKEND_PORT` 和 `NOFX_FRONTEND_PORT`。

### 依赖问题

如果遇到 Go 依赖问题，运行：
```bash
go mod tidy
```

如果遇到前端依赖问题，运行：
```bash
cd web
npm install
```

### 数据库连接问题

确保 `data` 目录存在并且有适当的读写权限。

## 高级开发选项

### 自定义数据库路径

```bash
go run main.go [自定义数据库路径]
```

### 开发模式运行

```bash
# 启用详细日志
go run main.go

# 在开发模式下，所有更改都会在重启后生效
```

## 部署到生产环境

当您完成本地开发和测试后，可以通过以下方式部署：

1. 将更改推送到 Git 仓库
2. 使用现有的 Docker 部署流程重新构建镜像
3. 或者使用 `start.sh` 脚本部署

## 注意事项

1. 本地开发模式不会持久化到 Docker 镜像中，需要重新构建才能将更改部署到 Docker 环境
2. 本地开发模式下的数据存储在本地文件系统中，与 Docker 容器的数据分离
3. 确保您的修改与 Docker 部署兼容，特别是在环境变量和路径方面