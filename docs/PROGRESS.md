# Progreso (bitácora de relevo entre agentes)

Regla: entradas breves (≤ 8 líneas por sesión). Lo más reciente arriba. Solo hechos útiles para el siguiente agente.

## Estado actual
- **Siguiente sesión:** S03
- **Bloqueos:** ninguno

## Historial
<!-- Plantilla:
### SXX — <título> — AAAA-MM-DD — ✅/⚠️
- Hecho:
- Decisiones/desvíos:
- Pendiente/deuda:
- Comandos de verificación:
-->

### S02 — Esqueleto backend Go — 2026-11-03 — ✅
- Hecho: módulo `tiquetera` (chi v5.2.1, pgx v5.7.2), config por env + `*_FILE`, slog JSON, middlewares (request-id, log, recover, security headers, body 64KB, timeout 10s), RFC7807, `/healthz` `/readyz`, graceful shutdown, subcomando `healthcheck` para distroless. Dockerfile multi-stage (19 MB, nonroot). api activa en compose (`127.0.0.1:8088`). CI: test -race, golangci-lint v2, govulncheck, trivy. `.gitignore` ampliado (secretos, firmas, builds móviles, backups) + `backend/.dockerignore` (whitelist).
- Decisiones: versiones a última estable verificada en fuentes oficiales (2026-11-03): Go 1.27.3, chi v5.5.0, pgx v5.12.0, testify v1.12.1, Postgres 18.6, distroless debian13, golangci-lint v2.15.0, govulncheck v1.9.0, actions checkout/setup-go v7. Dependabot cada 2 meses (cron), agrupado; ignora major de Postgres. Flutter estable actual: 3.47.9 (usar en S08 y activar ecosistema `pub` en dependabot.yml).
- Pendiente: S03 debe crear usuario de BD no-owner para la app (hoy usa owner). Go no instalado en host: `make test/lint/vuln` corren en Docker.
- Verificación: `make lint test` · `make up` → `curl 127.0.0.1:8088/readyz` 200.

### S01 — Monorepo e infraestructura local — 2026-11-03 — ✅
- Hecho: git (rama `feat/s01-infra`), .gitignore/.editorconfig/.gitattributes, Makefile (up/down/ps/logs/clean), infra/docker-compose.yml (db postgres:16.6-alpine healthy; api en profile `api`), infra/.env.example, CI (validación compose + gitleaks).
- Decisiones: Postgres expuesto solo en `127.0.0.1:5442` (5432/5433 ocupados por otros proyectos en el host).
- Pendiente: `make` no está instalado en Windows del host → usar Git Bash con make, o `docker compose --env-file infra/.env -f infra/docker-compose.yml ...`. S02 debe añadir jobs backend al CI y quitar el profile del api.
- Verificación: `docker compose --env-file infra/.env -f infra/docker-compose.yml ps` → db healthy.

### S00 — Documentación inicial — 2026-11-03 — ✅
- Hecho: CLAUDE.md, docs/, .claude/rules, .claude/skills.
