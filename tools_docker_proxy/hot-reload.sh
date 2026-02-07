#!/bin/sh
# 热更新监控脚本

# 安装inotify-tools
apk add --no-cache inotify-tools

# 启动应用
echo "Starting application..."
go run main.go &
APP_PID=$!

echo "Application started with PID $APP_PID"

# 监控文件变化并重启
inotifyd "echo 'File changed, restarting...'; kill $APP_PID; exit 0" . go &
INOTIFY_PID=$!

# 等待应用结束
wait $APP_PID