# 透明代理架构设计文档

## 1. 概述

透明代理功能是nofx系统的重要组成部分，旨在解决交易所API访问限制问题。该功能通过独立的代理服务实现，允许nofx在受限地区访问交易所API，同时保持系统架构的清晰和解耦。

## 2. 架构目标

- **解决访问限制**：绕过地区限制，访问受限的交易所API
- **保持透明性**：对nofx业务逻辑无影响
- **高可用性**：通过Docker容器化部署，确保服务稳定性
- **易维护性**：模块化设计，独立部署和维护
- **安全性**：保护敏感信息，防止泄露

## 3. 系统架构

```
+-------------------+        HTTP Request + X-Custom-API-URL Header        +------------------+        +------------------+
|                   | -------------------------------------------------> |                  | -----> |                  |
|    nofx Client    |                                                     |  Proxy Service   |        |  Exchange API    |
|                   | <-------------------------------------------------  |                  | <----- |                  |
+-------------------+         HTTP Response                              +------------------+        +------------------+
```

### 3.1 nofx Client (trader/binance_futures.go)

- **CustomTransport**: HTTP传输层拦截器
  - 拦截传出的HTTP请求
  - 将真实交易所URL放入 `X-Custom-API-URL` 请求头
  - 保持原有请求路径和参数不变

- **ProxyTraderWrapper** (trader/proxy_wrapper.go)
  - 代理模式管理器
  - 根据配置决定是否使用代理
  - 协调代理和直连两种模式

### 3.2 Proxy Service (tools_docker_proxy)

- **请求接收**: 接收来自nofx的请求
- **URL解析**: 从 `X-Custom-API-URL` 头部提取真实目标URL
- **请求转发**: 将请求转发到真实的目标地址
- **响应返回**: 将交易所响应返回给nofx
- **循环防护**: 防止请求被转发到自身

### 3.3 Exchange API

- 真实的交易所API服务
- 接收代理服务转发的请求
- 返回API响应

## 4. 核心组件

### 4.1 CustomTransport (trader/binance_futures.go)

```go
// CustomTransport 拦截HTTP请求并添加目标URL信息
type CustomTransport struct {
    http.RoundTripper
    TargetEndpoint string  // 真实交易所URL
}

func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    if ct.TargetEndpoint != "" {
        req.Header.Set("X-Custom-API-URL", ct.TargetEndpoint)
    }
    return ct.RoundTripper.RoundTrip(req)
}
```

### 4.2 代理服务 (tools_docker_proxy/main.go)

```go
// 从请求头获取目标API URL
func getTargetAPIURL(r *http.Request) string {
    return r.Header.Get("X-Custom-API-URL")
}

// 转发请求到真实目标
func forwardRequest(w http.ResponseWriter, r *http.Request, targetBaseURL string) {
    // 构建完整目标URL
    targetURL := targetBaseURL + r.URL.Path
    if r.URL.RawQuery != "" {
        targetURL += "?" + r.URL.RawQuery
    }
    // 转发请求
    // ...
}
```

## 5. 工作流程

1. **请求准备**：nofx内部使用真实交易所URL进行业务逻辑处理
2. **请求拦截**：CustomTransport拦截HTTP请求
3. **头部添加**：将真实交易所URL放入 `X-Custom-API-URL` 头部
4. **代理连接**：请求发送到代理服务 (`http://localhost:8081`)
5. **URL解析**：代理服务从头部提取真实目标URL
6. **请求转发**：代理服务将请求路径和参数附加到真实URL并转发
7. **响应返回**：代理服务将交易所响应返回给nofx

## 6. 配置管理

### 6.1 环境变量配置 (.env)

```bash
# 启用代理功能
USE_BINANCE_PROXY=true

# 代理服务端口
BINANCE_PROXY_PORT=8081

# 代理服务URL
BINANCE_PROXY_URL=http://localhost:8081
```

### 6.2 Docker部署配置

- 容器名称: `tools_docker_proxy`
- 端口: `8081`
- 支持热更新

## 7. 安全机制

### 7.1 循环防护

代理服务具有内置的循环防护机制，防止请求被转发到自身：

```go
// 检查是否试图转发到自身，防止循环
if strings.HasPrefix(targetBaseURL, ownAddress) {
    http.Error(w, "Prevented infinite loop: trying to forward to self", http.StatusBadRequest)
    return
}
```

### 7.2 敏感信息保护

- 真实交易所URL仅通过请求头传递，不在URL路径中暴露
- 代理服务不记录敏感信息（如签名、时间戳等）

## 8. 部署架构

### 8.1 Docker容器部署

- 代理服务运行在独立的Docker容器中
- 容器名称: `tools_docker_proxy`
- 端口映射: 主机8081 -> 容器8081
- 支持热更新

### 8.2 服务发现

- nofx通过 `http://localhost:8081` 连接代理服务
- 代理服务将请求转发到真实交易所URL

## 9. 故障处理

### 9.1 代理服务不可用

- nofx将无法连接到交易所API
- 需检查代理服务状态和网络连接

### 9.2 循环防护触发

- 当请求试图转发到代理服务自身时触发
- 返回400错误，防止无限循环

## 10. 监控和日志

### 10.1 代理服务日志

- 记录请求接收和转发信息
- 记录错误和异常情况
- 记录循环防护事件

### 10.2 nofx客户端日志

- 记录代理模式状态
- 记录请求发送和响应接收

## 11. 扩展性考虑

### 11.1 多交易所支持

架构支持多种交易所的代理访问，只需在nofx中配置不同的真实交易所URL。

### 11.2 高可用性

可通过部署多个代理服务实例和负载均衡实现高可用性。

## 12. 维护指南

### 12.1 服务更新

- 停止当前代理服务
- 拉取最新代码
- 重新构建Docker镜像
- 启动新服务

### 12.2 配置变更

- 修改环境变量后需重启nofx服务
- 修改代理服务配置需重启代理服务

## 13. 性能考虑

- 代理服务使用高效的HTTP转发机制
- 最小化额外延迟
- 支持连接池和并发处理