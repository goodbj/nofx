# 端口配置管理文档

本文档说明了NOFX项目中所有服务的端口配置及其管理方式。

## 端口配置概览

| 服务 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|------|
| 前端服务 | `FRONTEND_PORT` | 3300 | Web前端服务端口 |
| 后端服务 | `BACKEND_PORT` | 8888 | API后端服务端口 |
| 币安代理服务（外部） | `BINANCE_PROXY_PORT` | 8081 | 宿主机访问端口 |
| 币安代理服务（内部） | `BINANCE_PROXY_INTERNAL_PORT` | 8081 | 容器内部监听端口 |
| Guardian服务 | `GUARDIAN_PORT` | 8083 | Guardian监控服务端口 |
| 数据库 | `DB_PORT` | 5432 | PostgreSQL数据库端口 |
| Redis | `REDIS_PORT` | 6379 | Redis缓存端口 |

## 环境变量优先级

端口配置遵循以下优先级（从高到低）：

1. **环境变量** - 运行时通过环境变量设置的端口
2. **配置文件** - 代码中的默认配置值
3. **硬编码值** - 代码中的默认值（已废弃）

## 如何修改端口

### 方法一：使用配置脚本（推荐）

```bash
# 更新币安代理服务端口（外部端口 8090，内部端口 8091）
python scripts/update_ports.py 8090 8091
```

### 方法二：设置环境变量

```bash
# Linux/Mac
export BINANCE_PROXY_PORT=8090
export BINANCE_PROXY_INTERNAL_PORT=8081

# Windows
set BINANCE_PROXY_PORT=8090
set BINANCE_PROXY_INTERNAL_PORT=8081
```

### 方法三：修改配置文件

编辑 `config/ports.go` 文件中的默认端口值。

## 配置文件位置

- **Go配置**：`config/ports.go` - 集中管理所有端口配置
- **环境变量模板**：`config/port_env_template.go` - 支持环境变量覆盖
- **数据提供者**：`dataprovider/provider.go` - 使用集中配置
- **Docker配置**：`docker-compose.binance-proxy.yml` - 使用集中配置

## 注意事项

1. **端口冲突**：确保新端口不会与其他服务冲突
2. **防火墙**：确保新端口在防火墙中开放
3. **重启服务**：修改端口后需要重启相关服务
4. **容器重建**：修改容器内部端口需要重建Docker镜像

## 故障排除

### 服务无法启动
- 检查端口是否被占用：`netstat -an | grep <端口号>`
- 检查防火墙设置
- 查看服务日志获取具体错误信息

### 连接失败
- 确认服务是否在正确的端口监听
- 检查网络连通性
- 验证环境变量是否正确设置

## 最佳实践

1. **使用环境变量**：在生产环境中使用环境变量管理端口
2. **避免硬编码**：避免在代码中直接使用端口号
3. **文档同步**：修改端口后及时更新本文档
4. **测试验证**：修改端口后进行全面测试