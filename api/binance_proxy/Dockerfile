FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o binance-proxy .

FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建器阶段复制二进制文件
COPY --from=builder /app/binance-proxy .

# 暴露端口
EXPOSE 8082

# 运行二进制文件
CMD ["./binance-proxy"]