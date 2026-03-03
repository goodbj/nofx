# GO环境安装指南

##系统信息
- 操作系统: Windows 11 专业版 64位
-推荐GO版本: 1.21.5 或更高版本

##安步骤步骤

### 方法1：直接下载安装（推荐）

1. **下载GO安装包**
   -访问官方下载页面：https://go.dev/dl/
   - 下载适用于Windows的64位版本（go1.21.5.windows-amd64.msi）

2. **运行安装程序**
   -双下载的.msi文件
   -按照安装向导的提示进行安装
   - 默认安装路径通常是：C:\Program Files\Go\

3. **配置环境变量**
  安装程序通常会自动配置环境变量，但请确认以下设置：
   
   **系统环境变量需要添加：**
   -变量名：GOROOT
   - 变量值：C:\Program Files\Go
   
   **系统环境变量需要修改PATH：**
   - 添加：C:\Program Files\Go\bin
   - 添加：C:\Users\%USERNAME%\go\bin（如果存在）

4. **验证安装**
   重新打开命令提示符或PowerShell，运行：
   ```cmd
   go version
   go env
   ```

### 方法2：使用包管理器安装

如果您安装了Chocolatey包管理器：
```cmd
choco install golang
```

如果您安装了Scoop包管理器：
```cmd
scoop install go
```

##安装后验证

安装完成后，请重新启动命令行窗口，然后运行以下命令验证：

```cmd
go version
#应该显示类似：go version go1.21.5 windows/amd64

go env GOPATH
#应该显示您的GOPATH路径

go env GOROOT
# 应该显示GO安装路径
```

##常问题解决

### 1. 'go' 不是内部或外部命令
- 重新启动命令行窗口
-检查PATH环境变量是否包含GO的bin目录
- 确认GO安装路径

### 2.环境变量配置
如果自动配置失败，手动添加：
- GOROOT = C:\Program Files\Go
- GOPATH = C:\Users\%USERNAME%\go
- PATH += %GOROOT%\bin;%GOPATH%\bin

### 3.权限问题
- 以管理员身份运行安装程序
-确保安装目录有适当的读写权限

##安装完成后

GO环境安装完成后，您就可以运行NOFX项目了：

```cmd
#进项目目录
cd E:\AI\nofx_dev

#运行后端服务器
go run main.go

# 或者在tools_docker_proxy目录下运行透明代理
cd tools_docker_proxy
go run main.go
```

## 注意事项

1.安装GO后需要重启命令行窗口才能生效
2.确保防火墙不会阻止GO的网络访问
3. 如果使用代理，请配置相应的环境变量
4.使用最新稳定版本的GO

安装完成后，您就可以正常运行NOFX项目的后端服务和透明代理服务了。