# Progreso (bitácora de relevo entre agentes)

Regla: entradas breves (≤ 8 líneas por sesión). Lo más reciente arriba. Solo hechos útiles para el siguiente agente.

## Estado actual
- **Siguiente sesión:** S06
- **Bloqueos:** ninguno

## Historial
<!-- Plantilla:
### SXX — <título> — AAAA-MM-DD — ✅/⚠️
- Hecho:
- Decisiones/desvíos:
- Pendiente/deuda:
- Comandos de verificación:
-->

### S05 — Clientes, tiqueteras y QR — 2026-11-04 — ✅
- Hecho: `ClienteService` + `ClienteRepository` (crear en transacción, listar/buscar, detalle, PATCH, recarga, regenerar QR), `domain/cliente.go` (validaciones, `TodayInBogota`, `NewQRToken`, `HashQRToken`), audit_log, rutas admin.
- Decisiones: hash = SHA-256 del texto base64url tras `TQ1.` (S06 debe usar `domain.HashQRToken`). Bogotá = offset fijo -5 (sin DST ni tzdata en distroless). Códigos nuevos 422 `VALIDATION_ERROR`, 409 `TELEFONO_DUPLICADO`, 422 `CLIENTE_INACTIVO`; id mal formado = 404. Regenerar QR se permite a cliente inactivo.
- Negocio: la recarga suma el saldo anterior (incluso vencido); `almuerzos_comprados` de la nueva = recarga + saldo anterior, tope 500. S06: el débito debe filtrar `activa` para ordenarse con la recarga.
- Pendiente: S07: `GET /clientes/{id}/consumos` sigue en 501. Búsqueda `%q%` hace seq scan: aceptable con pocos clientes; `pg_trgm` si crece.
- Verificación: lint 0 issues; tests verdes; cobertura domain 97.4%, service 89.4%; `-race` en http/service/domain; E2E en Compose aislado (crear, duplicado, fecha vencida, buscar, PATCH null, recarga, regenerar, 0 tokens en logs, audit correcto).

### S04 — Autenticación y RBAC — 2026-11-04 — ✅
- Hecho: login/refresh/logout, `VerifyPassword`, JWT EdDSA (`internal/auth`), `AuthService` + repo con `WithTx`, bloqueo 5 fallos/15 min, rate limit 5/min IP+usuario, `authenticate` + `RequireRole`, audit_log de login/reuso/logout, `platform.Clock`, `http/errors.go`.
- Decisiones: `JWT_PRIVATE_KEY[_FILE]` = semilla Ed25519 base64 (`openssl rand -base64 32`), obligatoria. Inactivo/bloqueado/credencial mala: mismo 401 y siempre verifica hash (no revela estado). Máx. 2 verificaciones argon2 concurrentes (`mem_limit` 256m). Reuso solo audita si quedaban tokens activos (no alerta tras logout). `TRUST_PROXY_HEADERS` (default false) usa la última entrada de X-Forwarded-For. Deps: jwt v5.3.1, httprate v0.16.1 (aprobadas); indirectas xxh3 v1.0.2, cpuid v2.2.10.
- Pendiente: S05 reemplaza el placeholder 501 de `/v1/clientes` (ya protegido admin). S07: al desactivar usuario el access sigue válido hasta 15 min; el refresh ya se rechaza. Deploy prod: `TRUST_PROXY_HEADERS=true` tras Caddy. Rate limit en memoria: revisar si se escala a varias instancias. GO-2026-5932 sigue sin llamadas.
- Verificación: lint 0 issues; tests verdes; `-race` (imagen `golang:1.27.3`, no alpine) en http/service/postgres; cobertura domain 100%, service 82.9%; E2E en Compose aislado (login, 401/501, rotación, reuso revoca familia, logout, audit, 0 fugas en logs). `make` no existe en host: comandos equivalentes en Docker. El 429 y el 403 de cajero se probaron solo en tests HTTP.

### S03 — Migraciones, sqlc y OpenAPI — 2026-11-03 — ✅
- Hecho: migraciones 0001-0005 (7 tablas), rol `tiquetera_app` sin DELETE (la API ya no usa owner), sqlc en `internal/repository/postgres/sqlcgen`, queries base, `api/openapi.yaml`, `cmd/seed` idempotente, `platform.HashPassword`, targets `migrate`, `migrate-down`, `sqlc`, `seed`.
- Decisiones: restricciones extra no pedidas en DATA_MODEL: saldo <= comprados, FK compuesta consumos→tiqueteras(id, cliente_id), trigger append-only en consumos, CHECK de motivo. `audit_log.usuario_id` nullable (login fallido sin usuario). Grants explícitos: cada tabla nueva debe otorgarlos en su migración. PG18 devuelve `23001` (no `23503`) en FK RESTRICT. Deps nuevas aprobadas: uuid 1.6.0, x/crypto 0.58.0, migrate 4.20.1, testcontainers 0.44.0.
- Pendiente: S04 añade Verify y queries de refresh/bloqueo. S05: `EXPLAIN` de `SearchClientes`, escapar `%`/`_` de `q` y mapear `23505`. Sin `EXPLAIN` hecho en S03. GO-2026-5932 (x/crypto/openpgp) sin llamadas desde el código; subir x/crypto cuando haya release estable.
- Verificación: `make` no existe en host: comandos equivalentes en Docker (`make -n` validó targets). lint 0 issues, test verde, migrate up/down/up, seed 2 veces, `readyz` 200 con rol app. OpenAPI: redocly válido, 2 warnings (health sin 4xx).

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
