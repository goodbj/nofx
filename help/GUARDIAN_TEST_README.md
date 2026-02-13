# Guardian 浏览器自动化测试页面

## 概述

Guardian测试页面是一个专门用于测试Guardian浏览器自动化功能的界面。它允许您在不连接任何交易所的情况下，测试浏览器自动化功能是否正常工作。

## 功能特性

- **预设提示词**：提供常用的测试提示词，便于快速测试
- **自定义提示词**：允许输入自定义提示词进行个性化测试
- **浏览器自动化测试**：启动Chrome浏览器，自动打开AI网站并输入提示词
- **AI分析测试**：测试Guardian获取AI分析结果的功能
- **实时结果显示**：显示测试结果和状态信息

## 如何使用

### 1. 启动服务

使用以下命令启动后端服务：
```bash
cd e:/AI/nofx_Dev
go run main.go
```

使用以下命令启动前端服务：
```bash
cd e:/AI/nofx_Dev/web
npm install  # 如果是首次运行
npm install antd @ant-design/icons  # 安装UI组件库
npx vite
```

### 2. 访问测试页面

打开浏览器访问：
- 默认前端地址：http://localhost:5173/guardian-test
- 或者如果端口被占用，访问类似 http://localhost:5174/guardian-test 的地址

### 3. 执行测试

#### 浏览器自动化测试
1. 选择预设提示词或输入自定义提示词
2. 点击"测试Guardian浏览器自动化"按钮
3. 观察Chrome浏览器是否自动启动
4. 检查是否自动导航到AI网站并输入提示词

#### AI分析测试
1. 选择预设提示词或输入自定义提示词
2. 点击"测试AI分析功能"按钮
3. 查看返回的AI分析结果

## 技术细节

### 前端组件
- 文件路径：`web/src/pages/GuardianTestPage.tsx`
- 使用Ant Design构建用户界面
- 发送POST请求到后端API端点

### 后端API
- 测试浏览器自动化：`POST /api/test-guardian`
- 测试AI分析：`POST /api/test-guardian-analysis`
- 使用GuardianClient执行浏览器自动化

### GuardianClient
- 位于：`mcp/guardian_client.go`
- 使用chromedp库实现浏览器自动化
- 支持多种AI服务（DeepSeek、ChatGPT、Claude、Gemini等）
- 可配置显示模式（headless或可见）

## 故障排除

### Chrome浏览器未启动
- 确保系统已安装Chrome或Chromium浏览器
- 检查是否有足够的权限启动浏览器进程

### 提示词未正确输入
- 确认目标AI网站的DOM结构未发生变化
- 检查CSS选择器是否仍然有效

### 请求失败
- 验证后端服务是否正常运行
- 检查网络连接是否正常

## 注意事项

- 此测试不连接任何交易所，仅测试浏览器自动化功能
- 浏览器自动化功能需要本地安装Chrome/Chromium
- 在某些网络环境下可能需要额外的网络配置

## 维护

如需修改测试页面，编辑：
- 前端：`web/src/pages/GuardianTestPage.tsx`
- 后端：`api/server.go` 中的 `handleTestGuardian` 和 `handleTestGuardianAnalysis` 函数