# Docker环境下的Guardian浏览器自动化配置

## 问题描述
在Docker容器环境中，Chrome浏览器无法显示GUI窗口，导致Guardian浏览器自动化功能无法正常工作。

## 解决方案

### 1. Docker主机显示支持
要在Docker容器中显示GUI应用程序（如Chrome浏览器），需要配置X11转发或使用其他显示解决方案：

#### Windows WSL2 + Docker Desktop 方案：
1. 安装X11服务器（如VcXsrv或X410）
2. 在WSL2中配置DISPLAY环境变量：
   ```bash
   export DISPLAY=:0
   # 或者如果是通过网络连接的X11服务器
   export DISPLAY=$(cat /etc/resolv.conf | grep nameserver | awk '{print $2}'):0
   ```

#### Docker运行时添加显示权限：
```bash
# 运行容器时添加设备和环境变量
docker run -it --rm \
  -e DISPLAY=$DISPLAY \
  -v /tmp/.X11-unix:/tmp/.X11-unix \
  -v /dev/shm:/dev/shm \
  --name nofx-guardian \
  your-image-name
```

### 2. Guardian配置优化
为了解决Docker环境下的显示问题，Guardian已配置了以下Chrome启动参数：

- `--headless=false` - 禁用无头模式，启用GUI显示
- `--disable-gpu=false` - 启用GPU加速
- `--start-maximized=true` - 启动时最大化窗口
- `--no-sandbox=true` - 禁用沙盒（在容器中必要）
- `--disable-dev-shm-usage=true` - 禁用/ dev / shm使用
- `--remote-debugging-port=9222` - 启用远程调试

### 3. 登录和认证
- 在首次使用时，浏览器会弹出窗口，您可以在其中登录AI服务账户
- 为保持登录状态，Guardian配置了用户数据目录选项

### 4. 故障排除

#### 如果浏览器窗口仍未显示：
1. 确认主机上已安装并运行X11服务器
2. 检查DISPLAY环境变量设置是否正确
3. 确认Docker容器具有访问显示设备的权限

#### 如果AI服务要求登录：
1. 首次运行时，浏览器窗口会弹出
2. 手动在浏览器中登录所需的AI服务
3. 登录信息将被保存，后续运行无需重复登录

### 5. 使用方法

在前端配置中：
1. 将AI模型类型设置为"guardian"
2. 在Base URL字段中输入目标AI服务URL（如https://chat.deepseek.com/）
3. 系统将自动启动浏览器并导航到指定URL
4. 如果需要登录，浏览器窗口将显示供您手动登录