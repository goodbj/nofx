# NOFX 开发版 Docker 快速启动指南

## 一键部署

### 启动服务
双击运行 `start_nofx_dev_docker.bat` 文件，或在命令行中执行：
```cmd
.\start_nofx_dev_docker.bat
```

### 停止服务
双击运行 `stop_nofx_dev_docker.bat` 文件，或在命令行中执行：
```cmd
.\stop_nofx_dev_docker.bat
```

## 访问服务

- **前端界面**: http://localhost:3300
- **后端API**: http://localhost:8888
- **API健康检查**: http://localhost:8888/api/health

## 主要功能

✅ **前端服务** - 端口 3300  
✅ **后端服务** - 端口 8888  
✅ **Chrome无头模式** - 无界面浏览器操作  
✅ **热更新** - 代码更改即时生效  
✅ **数据持久化** - 数据库文件保存至本地  

## 文件结构

- `docker-compose.nofx-dev-chrome-hot.yml` - Docker Compose 配置文件
- `start_nofx_dev_docker.bat` - 启动脚本
- `stop_nofx_dev_docker.bat` - 停止脚本
- `data/data.db` - 数据库文件位置
- `DOCKER_DEPLOYMENT_README.md` - 详细部署文档

## 注意事项

1. 首次启动可能需要几分钟时间构建镜像
2. 确保端口 3300 和 8888 未被其他服务占用
3. 数据库文件会持久保存在 `E:\AI\nofx_dev\data\data.db`
4. 代码更改会自动触发热更新（后端）