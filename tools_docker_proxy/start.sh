#!/bin/sh
while true; do
  echo "Building application..."
  go build -o main .
  if [ -f main ]; then
    echo "Starting application..."
    ./main &
    PID=$!
    echo "Application started with PID $PID"
    # 监听.go文件变化
    inotifyd "kill $PID; exit 0" . go &
    wait $PID
  fi
done