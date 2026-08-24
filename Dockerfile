# ============================================
# 诗歌应用后端 - 多阶段构建 Dockerfile
# ============================================
# 构建: docker build -t poem-backend:latest .
# 运行: docker compose up -d

# ---------- Stage 1: 构建 ----------
FROM golang:1.26-alpine AS builder

# Alpine 国内镜像加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache git ca-certificates tzdata

# Go 模块代理（国内加速，可通过 build-arg 覆盖）
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /app

# 依赖缓存层
COPY go.mod go.sum ./
RUN go mod download

# 源码
COPY . .

# 静态编译
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /bin/server \
    cmd/server/main.go

# ---------- Stage 2: 运行 ----------
FROM alpine:3.20

# Alpine 国内镜像加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

# 时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /bin/server /app/server

# 迁移文件（容器内执行迁移用）
COPY migrations/ /app/migrations/

RUN chown -R appuser:appgroup /app

USER appuser

# 用户端 API :8080, 管理端 API :8081
EXPOSE 8080 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]
