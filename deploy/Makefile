.PHONY: docker-deploy docker-rebuild-agent docker-rebuild-backend docker-rebuild-frontend

COMPOSE_FILE := compose/docker-compose.local.yml
COMPOSE := $(shell if docker compose version >/dev/null 2>&1; then echo "docker compose"; else echo "docker-compose"; fi)

docker-deploy:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build

docker-rebuild-agent:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build modelcraft-agent

docker-rebuild-backend:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build backend

docker-rebuild-frontend:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d --build frontend
