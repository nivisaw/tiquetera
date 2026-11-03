# Reglas Docker / Infra

- Todo el tooling backend corre en contenedores: el host solo necesita Docker y Flutter SDK (el móvil no se dockeriza para compilar iOS; Android opcionalmente en CI con imagen `ghcr.io/cirruslabs/flutter:<versión exacta>`).
- Dockerfile API multi-stage: `golang:<última estable>-alpine` (build, `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`) → `gcr.io/distroless/static-debian13:nonroot`. Versión real: ver `backend/Dockerfile`.
- Postgres 18+: montar volumen en `/var/lib/postgresql` (no `/data`). Saltos de versión mayor = `pg_upgrade` planificado, nunca cambio directo de tag.
- Tags exactos de imágenes, nunca `latest`. `.dockerignore` estricto.
- Compose: healthchecks en db y api; `depends_on: condition: service_healthy`; redes internas; BD sin `ports` en prod; volúmenes nombrados; `restart: unless-stopped`; `security_opt: no-new-privileges`; `read_only: true` + `tmpfs` donde aplique; límites `mem_limit`/`cpus`.
- Archivos: `infra/docker-compose.yml` (dev), `infra/docker-compose.prod.yml` (override prod), `infra/Caddyfile`.
- En Windows no reescribir archivos con PowerShell 5.1 (`Get-Content`/`Set-Content` corrompen UTF-8): usar Edit/Write.
- Herramientas (migrate, sqlc, golangci-lint) se ejecutan vía `docker run` en el Makefile con versión fijada.
