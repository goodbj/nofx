# NOFX 开发版 Docker 部署指南

## 部署概述

此部署方案为 NOFX 开发版提供了完整的 Docker 容器化部署，包含以下特性：

- **前端服务**: 运行在端口 3300
- **后端服务**: 运行在端口 8888
- **Chrome 无头模式**: 支持自动化浏览器操作
- **热更新功能**: 代码更改时自动重新加载
- **数据持久化**: 数据库文件存储在本地目录

## 系统要求

- Docker Desktop 或 Docker Engine
- Docker Compose v2.x 或更高版本
- 至少 4GB 可用内存
- Windows/Linux/macOS

## 部署步骤

### 1. 启动服务

使用以下任一方式启动服务：

**方式一：使用批处理脚本（推荐）**
```cmd
start_nofx_dev_docker.bat
```

**方式二：使用 Docker Compose 命令**
```bash
docker-compose -f docker-compose.nofx-dev-chrome-hot.yml up --build -d
```

### 2. 验证部署

部署完成后，可以通过以下地址访问服务：

- **前端**: http://localhost:3300
- **后端 API**: http://localhost:8888
- **后端健康检查**: http://localhost:8888/api/health

### 3. 查看日志

实时查看服务日志：
```bash
docker-compose -f docker-compose.nofx-dev-chrome-hot.yml logs -f
```

## 服务架构

### 后端服务 (nofx-dev-backend)
- **端口**: 8888
- **功能**: 
  - 提供 RESTful API 接口
  - Chrome 无头模式支持
  - 热更新功能（使用 Air 工具）
  - 数据库连接和管理
- **数据卷映射**:
  - 项目源代码 → 容器内 `/app/` 目录（支持热更新）
  - 数据目录 → 容器内 `/app/data/` 目录（数据库持久化）

### 前端服务 (nofx-dev-frontend)
- **端口**: 3300 (宿主机) → 5173 (容器内)
- **功能**:
  - 提供 Web 用户界面
  - 与后端 API 通信
- **环境变量**:
  - `VITE_API_TARGET`: 指向后端服务地址

## 停止服务

### 使用批处理脚本（推荐）
```cmd
stop_nofx_dev_docker.bat
```

### 使用 Docker Compose 命令
```bash
docker-compose -f docker-compose.nofx-dev-chrome-hot.yml down
```

## 特殊功能说明

### Chrome 无头模式
- 已预配置无界面浏览器支持
- 所有浏览器操作将在后台运行，不会弹出窗口
- 适用于自动化测试和数据抓取

### 热更新机制
- 源代码修改后会自动重新编译和重启服务
- 缩短开发过程中的等待时间
- 支持 Go 和其他语言文件的监听

### 数据持久化
- SQLite 数据库文件保存在 `E:\AI\nofx_dev\data\data.db`
- 即使容器重启，数据也不会丢失
- 可随时备份该文件进行数据迁移

## 故障排除

### 1. 端口被占用
如果遇到端口冲突，请检查是否有其他服务占用了 3300 或 8888 端口。

### 2. 构建失败
首次构建可能需要较长时间，请耐心等待。如果持续失败，请检查：
- Docker 是否正常运行
- 网络连接是否正常
- 本地磁盘空间是否充足

### 3. Chrome 无头模式异常
如果遇到浏览器相关错误，检查环境变量设置是否正确：
- `CHROME_HEADLESS=true`
- `CHROME_BIN=/usr/bin/chromium-browser`

### 4. 热更新不生效
确认代码文件已正确挂载到容器中，且文件权限设置正确。

### 5. 获取余额失败
如果点击获取余额出现服务器错误，检查以下几点：
- 确认所有服务正常运行
- 确保 API 密钥配置正确

## 自定义配置

如需自定义配置，可修改 `docker-compose.nofx-dev-chrome-hot.yml` 文件：

- 更改端口映射
- 调整环境变量
- 修改数据卷路径
- 增加资源限制

## 维护建议

- 定期备份 `data.db` 数据库文件
- 监控容器资源使用情况
- 根据需要调整容器资源配置
- 保持 Docker 引擎更新至最新版本