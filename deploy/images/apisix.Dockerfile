FROM apache/apisix:3.9.0-debian

ARG APP_ENV=dev
ARG SERVICE_PORT=9080

USER root

ENV PORT=${SERVICE_PORT}

COPY deploy/cloudrun/apisix/docker-entrypoint.sh /docker-entrypoint-override.sh
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh
COPY apisix/config.yaml /usr/local/apisix/conf/config-template.yaml
COPY apisix/apisix.yaml /usr/local/apisix/conf/apisix-template.yaml
COPY deploy/configs/${APP_ENV}/apisix.yaml /etc/modelcraft/runtime.yaml

RUN chmod +x /docker-entrypoint-override.sh /usr/local/bin/load-flat-yaml-env.sh

EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=3 \
    CMD ["sh", "-c", "curl -sf http://localhost:${PORT}/apisix/prometheus/metrics || curl -sf http://localhost:${PORT}/"]

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/etc/modelcraft/runtime.yaml", "/docker-entrypoint-override.sh"]
