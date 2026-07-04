FROM mysql:8.0

ARG APP_ENV=dev

COPY deploy/configs/${APP_ENV}/mysql.yaml /etc/modelcraft/runtime.yaml
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh

RUN chmod +x /usr/local/bin/load-flat-yaml-env.sh

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/etc/modelcraft/runtime.yaml", "/usr/local/bin/docker-entrypoint.sh"]
CMD ["mysqld", "--default-authentication-plugin=mysql_native_password"]
