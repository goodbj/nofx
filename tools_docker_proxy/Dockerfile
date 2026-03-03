FROM golang:1.19-alpine

# 更换APK软件源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装依赖
RUN apk --no-cache add ca-certificates git

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o proxy-server .

# 暴露端口
EXPOSE 8081

# 运行应用
CMD ["./proxy-server"]