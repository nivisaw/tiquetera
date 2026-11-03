COMPOSE := docker compose --env-file infra/.env -f infra/docker-compose.yml
GO_IMAGE := golang:1.27.3-alpine
LINT_IMAGE := golangci/golangci-lint:v2.15.0-alpine
SQLC_IMAGE := sqlc/sqlc:1.31.1
GO_RUN := docker run --rm -v "$(CURDIR)/backend:/src" -v tiquetera-gomod:/go/pkg/mod -v tiquetera-gocache:/root/.cache -w /src -e GOTOOLCHAIN=local
TESTCONTAINERS := -v /var/run/docker.sock:/var/run/docker.sock --add-host=host.docker.internal:host-gateway -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal

export MSYS_NO_PATHCONV := 1

.PHONY: help env up down logs ps clean test lint vuln tidy migrate migrate-down sqlc seed

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

migrate: ## Aplica migraciones y asigna contraseña al rol de la app
	$(COMPOSE) run --rm migrate up
	$(COMPOSE) run --rm db-app-user

migrate-down: ## Revierte TODAS las migraciones
	$(COMPOSE) run --rm migrate down -all

sqlc: ## Genera código desde db/queries
	docker run --rm -v "$(CURDIR)/backend:/src" -w /src $(SQLC_IMAGE) generate

seed: ## Crea el admin inicial si no existe
	$(COMPOSE) run --rm --build seed

test: ## Tests backend con cobertura (requiere Docker para integración)
	$(GO_RUN) $(TESTCONTAINERS) $(GO_IMAGE) go test -cover ./...

lint: ## golangci-lint backend
	$(GO_RUN) $(LINT_IMAGE) golangci-lint run ./...

vuln: ## govulncheck backend
	$(GO_RUN) $(GO_IMAGE) sh -c "go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./..."

tidy: ## go mod tidy
	$(GO_RUN) $(GO_IMAGE) go mod tidy
