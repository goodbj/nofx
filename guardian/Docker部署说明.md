# Guardian Docker 部署说明

本文档介绍了如何将Guardian监控程序部署到Docker环境中，与现有的NoFx系统一起运行。

## 部署选项

### 1. 生产环境部署（含Guardian）

使用以下命令启动包含Guardian的生产环境：

```bash
docker-compose -f docker-compose.with-guardian.yml up -d
```

### 2. 开发环境部署（含Guardian）

使用以下命令启动包含Guardian的开发环境：

```bash
docker-compose -f docker-compose.dev.with-guardian.yml up -d
```

## 环境变量配置

在 `.env` 文件中添加以下Guardian相关的环境变量：

```bash
# Guardian 配置
GUARDIAN_API_KEY=your_api_key_here
GUARDIAN_AUTH_TOKEN=your_auth_token_here
GUARDIAN_AI_PROVIDER=deepseek
GUARDIAN_AI_ENDPOINT=https://chat.deepseek.com
GUARDIAN_LISTEN_INTERVAL=5000
```

## 服务说明

- `nofx`: NoFx后端服务
- `nofx-frontend`: NoFx前端服务  
- `nofx-guardian`: Guardian监控服务（生产环境）
- `nofx-guardian-dev`: Guardian监控服务（开发环境）

## 关键特性

1. **网络隔离**: 所有服务都在同一个Docker网络中，确保相互间通信
2. **依赖关系**: Guardian服务会在NoFx后端启动后再启动
3. **健康检查**: 每个服务都有健康检查机制
4. **数据持久化**: 数据卷用于持久化存储
5. **时区同步**: 容器与宿主机时间同步

## 注意事项

1. 确保 `.env` 文件中配置了正确的API密钥和其他认证信息
2. Guardian需要访问AI服务（如DeepSeek）的网络权限
3. Chrome浏览器自动化需要额外的系统资源
4. 如果需要自定义Chrome配置，可以通过环境变量调整

## 故障排除

- 查看服务日志：`docker logs <container_name>`
- 检查网络连接：`docker exec -it <container_name> ping nofx`
- 重启特定服务：`docker-compose -f <compose-file> restart nofx-guardian`