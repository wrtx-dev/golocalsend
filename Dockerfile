# 使用多阶段构建优化镜像大小
# 第一阶段：构建 Go 程序
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 将项目文件复制到容器
COPY . .

# 下载依赖并编译程序
RUN go mod tidy && go build

# 第二阶段：构建最小运行时镜像
FROM alpine:latest

# 添加运行时依赖
RUN apk add --no-cache ca-certificates

# 设置工作目录
WORKDIR /app

# 从构建阶段复制程序到运行时镜像
COPY --from=builder /app/golocalsend .

# 暴露服务端口（可选）
EXPOSE 53317

# 启动程序
CMD ["./golocalsend"]