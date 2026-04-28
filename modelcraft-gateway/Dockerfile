# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -ldflags="-w -s" -o modelcraft-gateway ./cmd/gateway/...

# Runtime stage
FROM alpine:latest

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/modelcraft-gateway ./modelcraft-gateway

RUN chmod +x /app/modelcraft-gateway && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8090

ENV GATEWAY_PORT=8090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/healthz || exit 1

CMD ["./modelcraft-gateway"]
