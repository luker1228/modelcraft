FROM phpmyadmin/phpmyadmin

ARG APP_ENV=dev

COPY deploy/configs/${APP_ENV}/phpmyadmin.yaml /etc/modelcraft/runtime.yaml
COPY deploy/scripts/load-flat-yaml-env.sh /usr/local/bin/load-flat-yaml-env.sh

RUN chmod +x /usr/local/bin/load-flat-yaml-env.sh

ENTRYPOINT ["/usr/local/bin/load-flat-yaml-env.sh", "/etc/modelcraft/runtime.yaml", "/docker-entrypoint.sh"]
