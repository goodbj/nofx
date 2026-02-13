# NOFX 本地开发环境设置

## 简介

此文档说明如何在本地环境中运行 NOFX 项目进行实时开发和测试，无需每次都构建 Docker 镜像。

## 快速开始

### Windows 用户

运行以下命令来设置和启动本地开发环境：

```cmd
local_dev_setup.bat
```

此脚本将：
1. 检查 Go、Node.js 和 npm 是否已安装
2. 创建 `.env` 配置文件（如不存在）
3. 创建 `data` 目录用于存储数据库
4. 安装 Go 和前端依赖
5. 同时启动后端和前端服务器

### 手动启动

如果您只想启动后端服务进行开发测试：

```cmd
local_run.bat
```

或者直接运行：

```bash
# 在项目根目录
go run main.go
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

## 特别提醒

由于您最近添加了动态止损、止盈和部分平仓功能，以及高级约束字段（如 `max_drawdown`, `min_target_profit`, `max_position_usd` 等），您可以在本地环境中：

1. 修改 [decision/engine.go](file:///e:/AI/nofx/decision/engine.go) 中的验证逻辑
2. 修改 [trader/binance_futures.go](file:///e:/AI/nofx/trader/binance_futures.go) 中的执行逻辑
3. 修改任何其他相关文件
4. 保存后重启服务进行实时测试

## 系统要求

- Go 1.19+
- Node.js 18+
- npm (随 Node.js 一起安装)

## 故障排除

如果遇到问题，请检查：

1. 确保已正确安装 Go、Node.js 和 npm
2. 确保端口 8888（后端）和 3300（前端）未被其他应用占用
3. 检查 `.env` 文件是否正确配置
4. 确保 `data` 目录存在且具有适当权限