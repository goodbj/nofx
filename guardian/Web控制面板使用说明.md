# Guardian Web 控制面板使用说明

这是一个图形化界面，用于控制Docker化部署的Guardian监测程序。

## 功能特性

- 🟢 一键启动包含Guardian的Docker服务
- 🔴 一键停止包含Guardian的Docker服务
- 🔍 检查服务运行状态
- 📋 查看Guardian实时日志
- 🌐 图形化界面操作
- 🖥️ 支持生产环境和开发环境切换

## 启动方法

### 1. 启动Web控制面板

```bash
# 方法一：使用批处理脚本
cd e:\AI\nofx_Dev\guardian
start_web_control_panel.bat

# 方法二：直接运行Go程序
cd e:\AI\nofx_Dev\guardian
go run web_controller.go
```

### 2. 访问控制面板

启动后，在浏览器中打开：
```
http://localhost:8085
```

## 使用步骤

1. **选择运行环境**：
   - 生产环境 (with-guardian)
   - 开发环境 (dev.with-guardian)

2. **使用控制按钮**：
   - **🟢 启动所有服务（含Guardian）**：启动包含Guardian的完整服务栈
   - **🔴 停止所有服务（含Guardian）**：停止所有相关服务
   - **🔍 检查服务状态**：查看当前所有服务的运行状态
   - **📋 查看Guardian日志**：获取Guardian服务的实时日志

3. **查看状态和日志**：
   - 状态信息会显示在页面顶部
   - 详细日志输出显示在页面底部

## 技术说明

- 后端：Go HTTP服务器，执行Docker命令
- 前端：纯HTML/CSS/JavaScript，无需额外依赖
- 通信：AJAX调用后端API执行Docker命令
- 安全：限制了文件访问路径，防止目录遍历攻击

## 注意事项

1. 确保系统已安装Docker和Docker Compose
2. 确保系统已安装Go运行环境
3. 确保Guardian相关的Docker Compose文件存在
4. 需要管理员权限执行Docker命令（取决于系统配置）
5. 服务器运行时会占用8085端口

## 故障排除

- 如果无法启动：检查Go环境和Docker是否正确安装
- 如果无法执行命令：检查Docker是否正在运行
- 如果页面无法加载：确认8085端口未被占用
- 如果命令执行失败：检查相关的Docker Compose文件是否存在

## 安全提醒

此控制面板允许执行Docker命令，仅应在可信的本地环境中使用，不应暴露在公共网络中。