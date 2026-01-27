# 配置本地Ollama服务器与DeepSeek模型

## 系统要求
- Ollama服务器正在运行（端口11434）
- 已安装`deepseek-r1:8b`模型
- NOFX开发版正在运行（端口8888/3300）

## 配置步骤

### 1. 验证Ollama服务
确认Ollama服务正在运行并可以访问：
```bash
curl http://127.0.0.1:11434/api/tags
```

### 2. 从前端界面配置
1. 打开浏览器访问 http://localhost:3300
2. 注册新账户或登录现有账户
3. 进入"AI模型配置"页面
4. 找到DeepSeek模型配置选项

### 3. 配置Ollama连接
在AI模型配置页面中，将Ollama模型配置为：

- **模型选择**: Ollama
- **启用模型**: 是
- **API端点**: http://host.docker.internal:11434
- **模型名称**: deepseek-r1:8b
- **API密钥**: 可留空（Ollama通常不需要API密钥）

### 4. 高级配置（如果需要）
如果系统要求使用OpenAI兼容的API格式，可能需要配置为：
- **API端点**: http://host.docker.internal:11434/v1
- **模型名称**: deepseek-r1:8b

### 5. 测试连接
配置完成后，系统应能通过以下方式访问本地模型：
- 容器内部使用 `host.docker.internal:11434` 访问主机上的Ollama服务
- 模型推理请求会被转发到本地运行的deepseek-r1:8b模型

## 注意事项
- Docker容器使用`host.docker.internal`来访问宿主机上的服务
- 确保防火墙允许11434端口的访问
- 本地模型推理可能需要较长时间，请耐心等待
- deepseek-r1:8b模型约5.2GB，需要足够的内存资源

## 故障排除
如果连接失败，请检查：
1. Ollama服务是否正在运行
2. deepseek-r1:8b模型是否已正确安装
3. Docker容器是否能访问宿主机
4. 网络防火墙设置