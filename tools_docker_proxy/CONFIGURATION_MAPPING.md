# 透明代理配置映射关系

## 主项目配置文件关联

透明代理服务与主项目的配置关系如下：

### 主项目 .env 配置（e:/AI/nofx_dev/.env）

```env
# Binance Proxy Configuration
USE_BINANCE_PROXY=true                    # 启用币安代理模式
BINANCE_PROXY_URL=http://localhost:8081   # 代理服务URL
BINANCE_PROXY_PORT=8081                   # 代理服务器端口
```

### 透明代理项目配置（当前项目）

- 端口: 8081 (与主项目配置一致)
- 服务名称: tools_docker_proxy
- 热更新: 已启用
- 功能: 透明代理服务，绕过地区限制访问受限制的交易所API

## 配置一致性说明

1. **端口一致性**：两个项目的端口都设置为8081，确保主项目能够正确调用代理服务
2. **代理启用**：主项目通过 `USE_BINANCE_PROXY=true` 启用代理功能
3. **URL匹配**：主项目中的 `BINANCE_PROXY_URL` 指向透明代理服务的地址

## 部署流程

1. 首先部署透明代理服务：
   ```bash
   cd tools_docker_proxy
   docker-compose -f docker-compose.hot.yml up -d --build
   ```

2. 然后启动主项目：
   ```bash
   cd ../
   go run main.go
   ```

## 故障排除

如果遇到连接问题，请检查：

1. 透明代理服务是否正在运行：
   ```bash
   docker ps --filter "name=tools_docker_proxy"
   ```

2. 端口8081是否可用：
   ```bash
   curl http://localhost:8081/health
   ```

3. 主项目配置是否正确：
   - `USE_BINANCE_PROXY=true`
   - `BINANCE_PROXY_URL=http://localhost:8081`
   - `BINANCE_PROXY_PORT=8081`

## 安全说明

透明代理服务遵循纯透传原则：
- 不处理任何业务逻辑
- 只做网络层的透明转发
- 不依赖环境变量进行控制
- 所有功能都通过原始交易者实现