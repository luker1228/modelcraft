FROM python:3.11-slim AS builder

WORKDIR /build
COPY modelcraft-agent/requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

FROM python:3.11-slim

ARG APP_ENV=dev
ARG SERVICE_PORT=8000

WORKDIR /app

COPY --from=builder /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH \
    PORT=${SERVICE_PORT}

COPY modelcraft-agent/ ./
COPY deploy/configs/${APP_ENV}/agent.yaml /app/config/runtime.yaml
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh

RUN chmod +x /usr/local/bin/load-flat-yaml-env.sh && \
    mkdir -p /app/logs /app/config

EXPOSE ${SERVICE_PORT}

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD python -c "import os, urllib.request; urllib.request.urlopen(f'http://localhost:{os.environ.get(\"PORT\", \"8000\")}/healthz')"

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/app/config/runtime.yaml"]
CMD ["sh", "-c", "python -m uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}"]
