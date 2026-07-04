# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git gcc musl-dev

# 设置 Go 环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 构建应用
RUN go build -ldflags="-w -s" -o main ./cmd/server

# 运行阶段
FROM alpine:latest

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用目录和数据目录
WORKDIR /app
RUN mkdir -p /app/data /app/configs /app/logs && \
    chown -R appuser:appgroup /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs/config.yaml ./configs/config.yaml

# 设置文件权限
RUN chmod +x main && \
    chown -R appuser:appgroup /app

# 切换到非 root 用户
USER appuser

# 暴露端口
# 端口通过环境变量 PORT 指定（默认 8080，云托管注入 80）
ARG PORT=8080
EXPOSE ${PORT}

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

# 设置环境变量
ENV GIN_MODE=release \
    CONFIG_FILE=/app/configs/config.yaml

# 运行应用
CMD ["./main", "-config", "/app/configs/config.yaml"]