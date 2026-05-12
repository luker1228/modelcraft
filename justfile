# justfile — ModelCraft monorepo root
#
# 本地 Docker 部署入口：
#   just deploy            # 启动（不重新构建）
#   just deploy force      # 构建后启动（代码有变动时使用）
#   just deploy init       # 从 example 生成 deploy/env/*.env
#   just deploy down       # 停止并删除容器
#   just deploy stop       # 停止容器
#   just deploy restart    # 重启容器
#   just deploy logs       # 查看所有服务日志
#   just deploy ps         # 查看状态
#   just deploy tools      # 启动附带 phpMyAdmin
#   just deploy build      # 仅构建镜像
#   just deploy rebuild    # 强制重新构建（无缓存）

set shell := ["bash", "-cu"]

compose_file := "deploy/compose/docker-compose.local.yml"
deploy_env_dir := "deploy/env"

[doc("Manage local docker deployment")]
deploy action="up":
    #!/usr/bin/env bash
    set -e

    COMPOSE_FILE="{{compose_file}}"
    ENV_DIR="{{deploy_env_dir}}"

    compose() {
        if docker compose version >/dev/null 2>&1; then
            docker compose -f "$COMPOSE_FILE" "$@"
        elif command -v docker-compose >/dev/null 2>&1; then
            docker-compose -f "$COMPOSE_FILE" "$@"
        else
            echo "❌ Neither 'docker compose' nor 'docker-compose' is available"
            exit 1
        fi
    }

    copy_if_missing() {
        local target="$1"
        local example="$2"
        if [ ! -f "$target" ]; then
            cp "$example" "$target"
            echo "✅ Created $target"
        fi
    }

    ensure_core_envs() {
        copy_if_missing "$ENV_DIR/mysql.local.env" "$ENV_DIR/mysql.local.env.example"
        copy_if_missing "$ENV_DIR/redis.local.env" "$ENV_DIR/redis.local.env.example"
        copy_if_missing "$ENV_DIR/backend.local.env" "$ENV_DIR/backend.local.env.example"
        copy_if_missing "$ENV_DIR/gateway.local.env" "$ENV_DIR/gateway.local.env.example"
        copy_if_missing "$ENV_DIR/frontend.local.env" "$ENV_DIR/frontend.local.env.example"
    }

    ensure_tools_env() {
        copy_if_missing "$ENV_DIR/phpmyadmin.local.env" "$ENV_DIR/phpmyadmin.local.env.example"
    }

    check_required_files() {
        local files=(
          "$ENV_DIR/mysql.local.env"
          "$ENV_DIR/redis.local.env"
          "$ENV_DIR/backend.local.env"
          "$ENV_DIR/gateway.local.env"
          "$ENV_DIR/frontend.local.env"
        )

        if [ "${1:-}" = "tools" ]; then
            files+=("$ENV_DIR/phpmyadmin.local.env")
        fi

        for file in "${files[@]}"; do
            if [ ! -f "$file" ]; then
                echo "❌ Missing: $file"
                exit 1
            fi
        done
    }

    print_urls() {
        echo "✅ Local deployment is up"
        echo "   Frontend:   http://localhost:3000"
        echo "   Gateway:    http://localhost:8090"
        echo "   Backend:    http://localhost:8080"
        echo "   MySQL:      localhost:6033"
    }

    case "{{action}}" in
        init)
            ensure_core_envs
            ensure_tools_env
            echo "✅ deploy/env/*.local.env initialized from examples"
            ;;
        up|start)
            ensure_core_envs
            check_required_files
            echo "🐳 Starting local deployment..."
            compose up -d
            echo ""
            compose ps
            echo ""
            print_urls
            ;;
        force)
            ensure_core_envs
            check_required_files
            echo "🐳 Building and starting local deployment..."
            compose up -d --build
            echo ""
            compose ps
            echo ""
            print_urls
            ;;
        down)
            echo "🛑 Stopping and removing containers..."
            compose down
            ;;
        stop)
            echo "⏸️  Stopping containers..."
            compose stop
            ;;
        restart)
            echo "🔄 Restarting containers..."
            compose restart
            compose ps
            ;;
        build)
            ensure_core_envs
            check_required_files
            echo "🏗️  Building images..."
            compose build
            ;;
        rebuild)
            ensure_core_envs
            check_required_files
            echo "🔄 Rebuilding images without cache..."
            compose build --no-cache
            ;;
        logs)
            compose logs -f
            ;;
        ps|status)
            compose ps
            ;;
        tools)
            ensure_core_envs
            ensure_tools_env
            check_required_files tools
            echo "🐳 Starting local deployment with phpMyAdmin..."
            compose --profile tools up -d
            compose ps
            echo ""
            echo "   phpMyAdmin: http://localhost:8081"
            ;;
        clean)
            echo "⚠️  This will remove containers and volumes."
            read -r -p "Are you sure? [y/N] " confirm
            if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
                compose down -v
                echo "✅ Containers and volumes removed"
            else
                echo "Cancelled"
            fi
            ;;
        *)
            echo "❌ Unknown action: {{action}}"
            echo "   Available: init, up, force, down, stop, restart, build, rebuild, logs, ps, tools, clean"
            exit 1
            ;;
    esac

[doc("Tail logs for one service or all services")]
logs service="":
    #!/usr/bin/env bash
    COMPOSE_FILE="{{compose_file}}"
    if docker compose version >/dev/null 2>&1; then
        if [ -z "{{service}}" ]; then
            docker compose -f "$COMPOSE_FILE" logs -f
        else
            docker compose -f "$COMPOSE_FILE" logs -f {{service}}
        fi
    elif command -v docker-compose >/dev/null 2>&1; then
        if [ -z "{{service}}" ]; then
            docker-compose -f "$COMPOSE_FILE" logs -f
        else
            docker-compose -f "$COMPOSE_FILE" logs -f {{service}}
        fi
    else
        echo "❌ Neither 'docker compose' nor 'docker-compose' is available"
        exit 1
    fi

[doc("Show local deployment status")]
ps:
    #!/usr/bin/env bash
    COMPOSE_FILE="{{compose_file}}"
    if docker compose version >/dev/null 2>&1; then
        docker compose -f "$COMPOSE_FILE" ps
    elif command -v docker-compose >/dev/null 2>&1; then
        docker-compose -f "$COMPOSE_FILE" ps
    else
        echo "❌ Neither 'docker compose' nor 'docker-compose' is available"
        exit 1
    fi

[doc("Start backend services only (mysql + backend + gateway)")]
deploy-backend action="up":
    #!/usr/bin/env bash
    set -e
    COMPOSE_FILE="{{compose_file}}"

    compose() {
        if docker compose version >/dev/null 2>&1; then
            docker compose -f "$COMPOSE_FILE" "$@"
        elif command -v docker-compose >/dev/null 2>&1; then
            docker-compose -f "$COMPOSE_FILE" "$@"
        else
            echo "❌ Neither 'docker compose' nor 'docker-compose' is available"
            exit 1
        fi
    }

    case "{{action}}" in
        up|start)
            echo "🐳 Starting backend services (mysql + backend + gateway)..."
            compose up -d modelcraft-mysql backend gateway
            echo ""
            compose ps
            echo ""
            echo "✅ Backend is up"
            echo "   Gateway:  http://localhost:8090"
            echo "   Backend:  http://localhost:8080"
            echo "   MySQL:    localhost:6033"
            ;;
        force)
            echo "🐳 Building and starting backend services..."
            compose up -d --build modelcraft-mysql backend gateway
            compose ps
            ;;
        down)
            compose stop backend gateway modelcraft-mysql
            echo "✅ Backend services stopped"
            ;;
        *)
            echo "Available: up, force, down"
            exit 1
            ;;
    esac
