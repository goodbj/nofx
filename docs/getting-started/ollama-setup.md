# Ollama 本地模型配置指南

## 概述

本指南介绍如何在 NOFX 系统中配置和使用本地 Ollama 模型。

## 功能说明

现在 NOFX 支持使用本地 Ollama 服务，您可以：

- 使用本地运行的大语言模型
- 无需互联网连接即可使用AI功能
- 自由选择各种开源模型
- 完全掌控数据隐私

## 系统要求

- Ollama 服务器正在运行（端口11434）
- 已安装所需的模型（如 llama3.1, deepseek-r1:8b 等）
- NOFX 系统正在运行

## 配置步骤

### 1. 验证 Ollama 服务

确认 Ollama 服务正在运行并可以访问：

```bash
curl http://127.0.0.1:11434/api/tags
```

### 2. 从前端界面配置

1. 打开浏览器访问 http://localhost:3300 (开发版) 或 http://localhost:3000 (稳定版)
2. 注册新账户或登录现有账户
3. 进入"AI模型配置"页面
4. 点击"添加模型"按钮
5. 在模型类型选择中，选择"Ollama"

### 3. 配置 Ollama 连接

在Ollama模型配置中填写以下信息：

- **模型选择**: Ollama
- **启用模型**: 是
- **API端点**: 
  - Docker环境: `http://host.docker.internal:11434`
  - 直接运行: `http://127.0.0.1:11434`
- **模型名称**: 您在Ollama中安装的模型名称（如 `llama3.1`, `deepseek-r1:8b` 等）
- **API密钥**: 可选，通常填 `ollama` 或留空

### 4. 高级配置

如果需要指定OpenAI兼容的API格式，可以配置为：
- **API端点**: `http://host.docker.internal:11434/v1`
- **模型名称**: 您的Ollama模型名称

## Docker 环境特殊说明

如果您在Docker环境中运行NOFX，需要注意：

- Docker容器内部的 `localhost` 指向容器本身，而不是主机
- 要访问主机上的Ollama服务，需要使用 `host.docker.internal`
- API端点应设置为 `http://host.docker.internal:11434`

## 创建使用Ollama的虚拟交易员

配置好Ollama模型后，您可以创建使用本地模型的虚拟交易员：

1. 进入"交易员管理"页面
2. 点击"创建新交易员"
3. 在AI模型选择中，选择您刚刚配置的Ollama模型
4. 配置其他交易参数
5. 启动交易员

## 故障排除

### 连接问题

如果连接失败，请检查：

1. **Ollama服务是否正在运行**
   ```bash
   curl http://127.0.0.1:11434/api/tags
   ```

2. **Docker网络连接**（如果使用Docker）
   - 确认API端点使用 `host.docker.internal:11434`
   - 检查防火墙设置

3. **模型名称是否正确**
   - 检查Ollama中是否存在指定的模型
   - 使用 `ollama list` 命令确认

4. **端口是否被占用**
   - 确认11434端口没有被其他服务占用

### 性能问题

- 本地模型推理可能需要较长时间，请耐心等待
- 确保有足够的内存资源分配给模型
- 复杂模型可能需要高性能GPU支持

## 支持的模型

Ollama支持多种开源模型，包括但不限于：

- Llama系列 (llama3.1, llama3.2等)
- Mixtral
- Mistral
- Gemma
- DeepSeek
- CodeLlama
- 以及其他GGUF格式的模型

您可以使用 `ollama pull <model-name>` 命令下载所需模型。

## 安全考虑

- 本地模型运行确保数据不会传输到外部服务器
- 所有处理都在本地完成，保护您的交易策略隐私
- 请注意，本地模型的准确性可能与云端商业模型有所不同