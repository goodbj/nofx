# AI自动化守护程序 (AI Automation Guardian)

AI自动化守护程序是一个独立的程序，用于实现NoFx系统与网页版AI服务之间的全自动交互。它可以：

1. 监听NoFx系统生成的提示词
2. 自动将提示词发送到网页版AI服务（如DeepSeek）
3. 自动获取AI的回复
4. 自动将AI决策提交回NoFx系统执行

## 架构设计

守护程序由以下几个核心模块组成：

- **监听模块**：监控NoFx系统生成的提示词
- **浏览器自动化模块**：控制网页版AI服务的输入输出
- **数据传输模块**：在NoFx和浏览器间安全传递数据
- **主控制器**：协调各模块的工作流程

## 安装和运行

### 1. 安装依赖

```bash
cd guardian
go mod tidy
```

### 2. 配置

守护程序支持通过环境变量进行配置：

- `NOFX_API_ENDPOINT`: NoFx API端点地址（默认：http://localhost:8888/api）
- `GUARDIAN_API_KEY`: API密钥
- `GUARDIAN_AUTH_TOKEN`: 认证令牌
- `GUARDIAN_AI_PROVIDER`: AI服务提供商（默认：deepseek）
- `GUARDIAN_AI_ENDPOINT`: AI服务端点（默认：https://chat.deepseek.com）

### 3. 运行

```bash
go run main.go
```

或者构建后运行：

```bash
go build -o guardian main.go
./guardian
```

## 工作流程

1. 守护程序启动后，开始监听NoFx系统生成的提示词
2. 当检测到新的提示词时，将其发送到配置的AI服务网页
3. 自动等待并获取AI的响应
4. 将AI的决策结果发送回NoFx系统执行
5. 整个过程无需人工干预，实现完全自动化

## 配置选项

守护程序的详细配置可以在 `config/guardian_config.go` 中查看和修改，包括：

- 浏览器配置（路径、窗口大小、代理等）
- 数据传输配置（超时、重试等）
- 监听间隔
- 日志级别

## 安全考虑

- 所有数据传输都通过加密通道进行
- 支持API密钥和认证令牌
- 不在本地存储敏感信息

## 错误处理和重试

守护程序内置了完善的错误处理和重试机制：

- 网络请求失败时自动重试
- 浏览器操作失败时自动重试
- 数据传输失败时自动重试

## 监控和日志

守护程序会记录详细的运行日志，包括：

- 提示词接收和发送
- AI响应获取
- 决策执行结果
- 错误和异常信息

## 注意事项

- 确保运行环境中已安装Chrome或Chromium浏览器
- 确保网络可以访问配置的AI服务
- 遵守AI服务的使用条款和频率限制