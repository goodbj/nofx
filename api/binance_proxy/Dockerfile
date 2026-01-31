# 使用官方Go镜像作为构建环境
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 配置Go模块代理
ENV GOPROXY=direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 安装git和其他必要工具
RUN apk add --no-cache git

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 检查是否需要热更新功能
ARG HOT_RELOAD=false
ENV HOT_RELOAD=${HOT_RELOAD}

# 如果启用热更新，则安装air
RUN if [ "$HOT_RELOAD" = "true" ] ; then \
    go install github.com/air-verse/air@latest && \
    apk add --no-cache git ; \
    fi

# 构建应用（仅在非热更新模式下）
RUN if [ "$HOT_RELOAD" != "true" ] ; then \
    CGO_ENABLED=0 GOOS=linux go build -o binance-proxy . ; \
    fi

# 使用轻量级基础镜像运行应用
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制整个应用目录（用于热更新）
COPY --from=builder /app .

# 检查是否需要热更新
ARG HOT_RELOAD=false
ENV HOT_RELOAD=${HOT_RELOAD}

# 暴露端口
EXPOSE 8082

# 创建启动脚本
RUN echo '#!/bin/sh\nif [ "$HOT_RELOAD" = "true" ] ; then\n  cd /app && cp /app/.air.conf . && go mod download && air -c .air.conf\nelse\n  ./binance-proxy\nfi' > /start.sh && chmod +x /start.sh

# 运行应用 - 根据是否启用热更新选择不同的命令
CMD ["/start.sh"]