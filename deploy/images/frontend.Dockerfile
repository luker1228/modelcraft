FROM node:20-alpine AS deps

WORKDIR /app

COPY modelcraft-front/package.json modelcraft-front/package-lock.json ./
RUN npm ci

FROM node:20-alpine AS builder

ARG APP_ENV=dev

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY modelcraft-front/ ./
COPY deploy/configs/${APP_ENV}/frontend.yaml /tmp/runtime.yaml

RUN BACKEND_URL="$(awk -F': ' '/^BACKEND_URL:/ {print $2}' /tmp/runtime.yaml | sed 's/^"//; s/"$//')" && \
    PORT="$(awk -F': ' '/^PORT:/ {print $2}' /tmp/runtime.yaml | sed 's/^"//; s/"$//')" && \
    export BACKEND_URL PORT && \
    npm run build

FROM node:20-alpine AS runner

ARG APP_ENV=dev
ARG SERVICE_PORT=3000

WORKDIR /app

ENV NODE_ENV=production \
    TZ=Asia/Shanghai \
    PORT=${SERVICE_PORT} \
    HOSTNAME=0.0.0.0

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder --chown=appuser:appgroup /app/public ./public
COPY --from=builder --chown=appuser:appgroup /app/.next/standalone ./
COPY --from=builder --chown=appuser:appgroup /app/.next/static ./.next/static
COPY deploy/configs/${APP_ENV}/frontend.yaml /app/config/runtime.yaml
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh

RUN chmod +x /usr/local/bin/load-flat-yaml-env.sh

USER appuser

EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/ || exit 1

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/app/config/runtime.yaml"]
CMD ["node", "server.js"]
