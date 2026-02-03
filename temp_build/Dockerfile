# 使用官方Go镜像作为构建环境
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装git
RUN apk add --no-cache git

# 复制go module文件
COPY api/binance_proxy/go.mod api/binance_proxy/go.sum ./
RUN go mod download

# 复制源码
COPY api/binance_proxy/ ./
COPY logger_module/ ../logger_module/

# 修改go.mod以适应容器环境
RUN sed -i 's|../../logger_module|../logger_module|g' go.mod

# 下载依赖
RUN go mod tidy

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o binance-proxy .

# 使用轻量级基础镜像运行应用
FROM alpine:latest

# 安装curl用于健康检查（可选）
RUN apk add --no-cache curl

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/binance-proxy .

# 暴露端口
EXPOSE 8082

# 运行应用
CMD ["./binance-proxy"]