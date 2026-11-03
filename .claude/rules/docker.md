# Reglas Docker / Infra

- Todo el tooling backend corre en contenedores: el host solo necesita Docker y Flutter SDK (el móvil no se dockeriza para compilar iOS; Android opcionalmente en CI con imagen `ghcr.io/cirruslabs/flutter:<versión exacta>`).
- Dockerfile API multi-stage: `golang:1.23-alpine` (build, `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) → `gcr.io/distroless/static-debian12:nonroot`.
- Tags exactos de imágenes, nunca `latest`. `.dockerignore` estricto.
- Compose: healthchecks en db y api; `depends_on: condition: service_healthy`; redes internas; BD sin `ports` en prod; volúmenes nombrados; `restart: unless-stopped`; `security_opt: no-new-privileges`; `read_only: true` + `tmpfs` donde aplique; límites `mem_limit`/`cpus`.
- Archivos: `infra/docker-compose.yml` (dev), `infra/docker-compose.prod.yml` (override prod), `infra/Caddyfile`.
- Herramientas (migrate, sqlc, golangci-lint) se ejecutan vía `docker run` en el Makefile con versión fijada.
