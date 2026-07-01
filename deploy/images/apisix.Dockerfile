FROM apache/apisix:3.9.0-debian

ARG APP_ENV=dev
ARG SERVICE_PORT=9080

USER root

ENV PORT=${SERVICE_PORT}

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl python3 && \
    rm -rf /var/lib/apt/lists/*

COPY apisix/docker-entrypoint.sh /docker-entrypoint-override.sh
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh
RUN mkdir -p /app/apisix/lua /etc/apisix
COPY apisix/config.template.yaml /app/apisix/config.template.yaml
COPY apisix/apisix.template.yaml /app/apisix/apisix.template.yaml
COPY apisix/lua /app/apisix/lua
COPY deploy/configs/${APP_ENV}/apisix.yaml /app/apisix/runtime.yaml

RUN chmod +x /docker-entrypoint-override.sh /usr/local/bin/load-flat-yaml-env.sh

EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=3 \
    CMD ["sh", "-c", "curl -sf http://localhost:${PORT}/health >/dev/null"]

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/app/apisix/runtime.yaml", "/docker-entrypoint-override.sh"]
