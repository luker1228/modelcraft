#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <dev|online|cloudrun> <compose-subcommand> [args...]" >&2
  exit 1
fi

DEPLOY_ENV="$1"
shift

COMPOSE_BIN="docker compose"
if ! docker compose version >/dev/null 2>&1; then
  COMPOSE_BIN="docker-compose"
fi

export APP_ENV="$DEPLOY_ENV"
export BACKEND_HOST_PORT=8080
export APISIX_HOST_PORT=9080
export AGENT_HOST_PORT=8000
export FRONTEND_HOST_PORT=3100
export MYSQL_HOST_PORT=6033
export PHPMYADMIN_HOST_PORT=8081

case "$DEPLOY_ENV" in
  dev)
    export BACKEND_CONTAINER_PORT=8080
    export APISIX_CONTAINER_PORT=9080
    export AGENT_CONTAINER_PORT=8000
    export FRONTEND_CONTAINER_PORT=3000
    profiles=(--profile dev)
    default_services=(backend apisix agent frontend mysql)
    ;;
  online)
    export BACKEND_CONTAINER_PORT=8080
    export APISIX_CONTAINER_PORT=9080
    export AGENT_CONTAINER_PORT=8000
    export FRONTEND_CONTAINER_PORT=3000
    profiles=()
    default_services=(backend apisix agent frontend)
    ;;
  cloudrun)
    export BACKEND_CONTAINER_PORT=8080
    export APISIX_CONTAINER_PORT=9080
    export AGENT_CONTAINER_PORT=8085
    export FRONTEND_CONTAINER_PORT=80
    profiles=()
    default_services=(backend apisix agent frontend)
    ;;
  *)
    echo "unsupported env: $DEPLOY_ENV" >&2
    exit 1
    ;;
esac

subcommand="$1"
shift

compose_cmd=($COMPOSE_BIN -f docker-compose.yml "${profiles[@]}")

for arg in "$@"; do
  if [ "$arg" = "phpmyadmin" ]; then
    compose_cmd=($COMPOSE_BIN -f docker-compose.yml "${profiles[@]}" --profile tools)
    break
  fi
done

case "$subcommand" in
  up)
    if [ "$#" -eq 0 ]; then
      exec "${compose_cmd[@]}" up -d "${default_services[@]}"
    fi
    exec "${compose_cmd[@]}" up "$@"
    ;;
  build)
    if [ "$#" -eq 0 ]; then
      exec "${compose_cmd[@]}" build "${default_services[@]}"
    fi
    exec "${compose_cmd[@]}" build "$@"
    ;;
  config|down|ps|logs|restart|pull)
    exec "${compose_cmd[@]}" "$subcommand" "$@"
    ;;
  *)
    exec "${compose_cmd[@]}" "$subcommand" "$@"
    ;;
esac
