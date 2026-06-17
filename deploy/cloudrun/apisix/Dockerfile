# APISIX 云托管镜像
# 基于官方 APISIX 镜像，注入项目配置并支持 $PORT 动态注入
FROM apache/apisix:3.9.0-debian

USER root

# 复制启动脚本（在 APISIX 启动前用环境变量替换端口）
COPY docker-entrypoint.sh /docker-entrypoint-override.sh
RUN chmod +x /docker-entrypoint-override.sh

# 复制项目配置文件
COPY config.yaml /usr/local/apisix/conf/config-template.yaml
COPY apisix.yaml /usr/local/apisix/conf/apisix.yaml

# 端口通过环境变量 PORT 指定（默认 9080，云托管注入 80）
ARG PORT=9080
EXPOSE ${PORT}

HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=3 \
    CMD ["sh", "-c", "curl -sf http://localhost:${PORT:-9080}/apisix/prometheus/metrics || curl -sf http://localhost:${PORT:-9080}/"]

# 使用自定义入口脚本，将原始 CMD 作为参数传递
ENTRYPOINT ["/docker-entrypoint-override.sh"]
CMD ["docker-entrypoint.sh"]
