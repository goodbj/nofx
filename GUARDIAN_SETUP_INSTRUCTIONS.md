# Guardian 功能配置说明

## 环境变量配置

在 `.env` 文件中已正确配置了以下 Guardian 相关的环境变量：

### 基础配置
```bash
# 启用显示模式（非Docker环境下）
GUARDIAN_DISPLAY_ENABLED=true

# 启用手动模式（浏览器保持长时间打开供用户登录）
GUARDIAN_MANUAL_MODE=true

# 自动模式下的浏览器超时时间（秒）
GUARDIAN_BROWSER_TIMEOUT_SECONDS=290  # 4分50秒，避免与下个5分钟查询启动冲突

# 手动模式下的浏览器超时时间（秒）
GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS=1200  # 20分钟

# 长时间浏览器窗口超时时间（秒）
GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS=1800  # 30分钟

# Chrome用户数据目录（用于保存Cookie和登录状态）
GUARDIAN_USER_DATA_DIR=./chrome_profile

# 自动模式下保持浏览器打开（用于观察结果）
GUARDIAN_AUTO_KEEP_OPEN=true

# 自动模式下保持浏览器打开的时间（秒）
GUARDIAN_AUTO_KEEP_OPEN_SECONDS=30
```

### DeepSeek 专用配置
```bash
# DeepSeek 目标URL
GUARDIAN_TARGET_URL=https://chat.deepseek.com/

# DeepSeek 页面元素选择器
GUARDIAN_INPUT_SELECTOR=textarea[placeholder='给 DeepSeek 发送消息']
GUARDIAN_BUTTON_SELECTOR=div._7436101.ds-icon-button.ds-icon-button--l.ds-icon-button--sizing-container[role='button']
GUARDIAN_RESPONSE_SELECTOR=div.ds-message._63c77b1
```

## 如何使用

### 1. 手动登录模式
当设置 `GUARDIAN_MANUAL_MODE=true` 时：
- 浏览器将保持打开长达 20 分钟（可根据 `GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS` 调整）
- 用户有充足时间完成 DeepSeek 账户登录
- 登录成功后，后续 AI 请求将自动使用已登录的会话
- 登录状态将保存在用户数据目录中，下次启动时无需重新登录

### 2. 自动模式
当设置 `GUARDIAN_MANUAL_MODE=false` 时：
- 浏览器将自动执行 AI 请求
- 超时时间为 60 秒（可根据 `GUARDIAN_BROWSER_TIMEOUT_SECONDS` 调整）

## 会话保持机制

Guardian 现在支持浏览器会话保持功能，登录状态将被持久化：

- 通过 `GUARDIAN_USER_DATA_DIR` 环境变量指定用户数据目录
- Chrome 的配置、Cookie 和登录状态都将保存在此目录中
- 重启应用后，浏览器会恢复之前的登录状态
- 默认目录为 `./chrome_profile`

## 故障排除

### 1. 浏览器未弹出
- 确认 `GUARDIAN_DISPLAY_ENABLED=true`
- 确认未设置 `DOCKER_ENV` 或将其设为空值
- 确认系统上安装了 Chrome 浏览器

### 2. 浏览器打开是空白页面
- 这通常是由于Chrome的安全限制或页面加载时间不足造成的
- 已经移除了过多的安全限制参数，允许正常的网站访问
- 已经增加了页面加载等待时间到5秒，确保页面能够正确加载（适用于VPN环境）
- 已经添加了更多浏览器兼容性参数，提高网站加载成功率
- 如果仍有问题，请检查网络连接是否正常
- 确保可以正常访问目标AI服务网站（如 https://chat.deepseek.com）

### 2. 浏览器打开是空白页面或启动失败
- 这通常是由于Chrome的安全限制、端口冲突或页面加载时间不足造成的
- 已经移除了过多的安全限制参数，允许正常的网站访问
- 已经增加了页面加载等待时间到5秒，确保页面能够正确加载（适用于VPN环境）
- 已经添加了更多浏览器兼容性参数，提高网站加载成功率
- 已将远程调试端口设置为随机端口(0)，避免端口冲突问题
- 已将浏览器窗口大小固定为550x850，适合AI聊天界面的最佳显示效果
- 如果仍有问题，请检查网络连接是否正常
- 确保可以正常访问目标AI服务网站（如 https://chat.deepseek.com）

### 3. 自动填充和响应获取
- Guardian现在会更可靠地自动填充提示词到输入框
- 改进了输入框清空和内容发送逻辑
- 改进了按钮点击逻辑，增加了滚动到元素功能
- 增强了响应提取机制，有多种备用方案获取AI响应
- 特别为DeepSeek添加了专用选择器（输入框和按钮），提高元素定位准确性
- 利用DeepSeek按钮状态变化（发送图标→等待图标）检测AI处理进度，提高响应获取准确性
- 在获取响应后尝试自动点击复制按钮，方便内容复用
- 如果主要响应提取失败，会尝试多种备用方法获取内容

### 4. 首次登录设置
- 首次使用Guardian时，需要先完成AI服务网站的登录
- 点击“设置浏览器”按钮打开长时间浏览器窗口
- 在打开的浏览器中完成账户登录
- 登录状态会被保存在用户数据目录中，后续调用将自动使用已登录的会话
- 登录成功后，可以关闭浏览器窗口，后续的自动化调用将使用已保存的登录状态

### 5. 通过按钮打开长时间浏览器窗口
- 新增了专门的API端点 `/open-guardian-browser` 来打开长时间保持的浏览器窗口
- 可以通过点击按钮的方式打开一个长时间保持的浏览器窗口，方便进行登录等操作
- 使用 `GUARDIAN_LONG_BROWSER_TIMEOUT_SECONDS` 控制浏览器窗口保持时间（默认1800秒/30分钟）
- 程序会自动增加页面加载和响应等待时间，以确保VPN环境下的正常运行
- 浏览器现在会等待AI响应生成完成后再关闭，确保输入和输出都能正确处理

### 4. 自动模式下浏览器快速关闭
- 在自动模式下，浏览器最多会在5分钟内完成任务并关闭（无论任务是否完成）
- 如果需要在自动模式下观察结果，可以设置 `GUARDIAN_AUTO_KEEP_OPEN=true`
- 使用 `GUARDIAN_AUTO_KEEP_OPEN_SECONDS` 控制保持打开的时间（默认30秒）
- 如需更长时间保持浏览器打开（如进行登录操作），请设置 `GUARDIAN_MANUAL_MODE=true` 并配合 `GUARDIAN_MANUAL_BROWSER_TIMEOUT_SECONDS=1200`
- 程序会自动增加页面加载和响应等待时间，以确保VPN环境下的正常运行
- 浏览器现在会等待AI响应生成完成后再关闭（最长不超过5分钟），确保输入和输出都能正确处理

### 4. 登录后无法正常使用
- 确保登录完成后等待一段时间再发起 AI 请求
- 检查 Cookie 和会话是否保持有效

### 3. 超时时间未生效
- 检查环境变量拼写是否正确
- 确认后端服务重启以加载新的环境变量

## 运行服务

要运行后端服务，请确保：
1. 已安装 Go 语言环境
2. 在项目根目录运行：`go run main.go`
3. 或者使用预编译的二进制文件（如果为正确平台编译）

## 前端测试

前端提供了 Guardian 测试页面，可以通过以下路径访问：
- http://localhost:3300/guardian-test