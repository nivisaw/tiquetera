COMPOSE := docker compose --env-file infra/.env -f infra/docker-compose.yml

.PHONY: help env up down logs ps clean

help: ## Lista comandos
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

env: ## Crea infra/.env desde el ejemplo si no existe
	@test -f infra/.env || cp infra/.env.example infra/.env

up: env ## Levanta el stack local
	$(COMPOSE) up -d --wait

down: ## Detiene el stack
	$(COMPOSE) down

logs: ## Logs del stack
	$(COMPOSE) logs -f --tail=100

ps: ## Estado de servicios
	$(COMPOSE) ps

clean: ## Detiene y BORRA volúmenes (datos locales)
	$(COMPOSE) down -v
