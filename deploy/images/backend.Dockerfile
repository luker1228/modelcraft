FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

COPY modelcraft-backend/go.mod modelcraft-backend/go.sum ./
RUN go mod download && go mod verify

COPY modelcraft-backend/ ./
RUN go build -ldflags="-w -s" -o main ./cmd/server

FROM alpine:latest

ARG APP_ENV=dev
ARG SERVICE_PORT=8080

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

ENV TZ=Asia/Shanghai \
    GIN_MODE=release \
    CONFIG_FILE=/app/configs/config.yaml \
    PORT=${SERVICE_PORT}

WORKDIR /app
RUN mkdir -p /app/data /app/configs /app/logs && \
    chown -R appuser:appgroup /app

COPY --from=builder /app/main .
COPY deploy/configs/${APP_ENV}/backend.yaml ./configs/config.yaml

RUN chmod +x main && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

CMD ["./main", "-config", "/app/configs/config.yaml"]
