# Progreso (bitácora de relevo entre agentes)

Regla: entradas breves (≤ 8 líneas por sesión). Lo más reciente arriba. Solo hechos útiles para el siguiente agente.

## Estado actual
- **Siguiente sesión:** S10
- **Bloqueos:** ninguno. iOS en pausa (sin Mac): verificar en Fase 3 con GitHub Actions macOS (simulador) o Codemagic/TestFlight

## Historial
<!-- Plantilla:
### SXX — <título> — AAAA-MM-DD — ✅/⚠️
- Hecho:
- Decisiones/desvíos:
- Pendiente/deuda:
- Comandos de verificación:
-->

### S09 — Login y sesión por rol — 2026-11-10 — ✅
- Hecho: `features/auth` (`User` freezed con `Role`, `AuthRepository` login/restore/logout, `AuthController` keepAlive), `LoginPage` (validación, cargando, errores 401/429/sin red/timeout), `authRedirect` puro + `refreshListenable` (admin → `/clientes`, cajero → `/escaner`, placeholders con logout). Si el refresh falla, `sessionExpiredProvider` lleva a login.
- Decisiones: el usuario se guarda en secure storage junto a los tokens (el refresh no devuelve user). Rol desconocido = sin sesión. Logout borra la sesión local primero y revoca en servidor si hay red. `API_BASE_URL` debe incluir `/v1` (ej. `http://10.0.2.2:8088/v1`).
- Android: `INTERNET` en el manifest main (faltaba en release). `network_security_config` solo en debug: permite http a 10.0.2.2/localhost.
- Verificación: analyze 0; 24/24 tests; cobertura auth data 96.7%, domain 100%, presentation 93%. E2E en emulador con Compose aislado: login admin → Clientes, sesión restaurada tras reiniciar, logout 204 → login; 0 tokens/contraseñas en logs de la API.

### S08 — Esqueleto Flutter — 2026-11-10 — ✅
- Hecho: `mobile/` (`co.marvinapps.tiquetera`, Android/iOS), Flutter 3.47.9 / Dart 3.13.7 en `%USERPROFILE%/dev/flutter`. Deps fijadas exactas. `core/` con config, theme, router, network, storage y providers; `features/home` placeholder; l10n ARB es (`lib/l10n/gen`); `very_good_analysis` 11.0.0.
- Red: dio con timeouts de 10 s; `AuthInterceptor` usa un único refresh compartido (future) para 401 concurrentes, reintenta una vez y, si el refresh devuelve 401, limpia tokens y avisa `sessionExpiredProvider`. `ErrorMappingInterceptor` traduce RFC7807 a la `AppFailure` sellada.
- Config: `--dart-define=API_BASE_URL=... FLAVOR=dev|prod`; prod exige https; falla rápido si falta. Android: `allowBackup=false`, `usesCleartextTraffic=false`.
- Decisiones: Dart 3.13 usa la sintaxis de constructor `new(...)` (lint `unnecessary_type_name_in_constructor`). Archivos `.g.dart` commiteados.
- Android: arranca en AVD `Medium_Phone_API_37.0` (API 37). Con Impeller/OpenGLES del emulador la pantalla queda negra; con `am start ... --ez enable-impeller false` (Skia) renderiza bien. Es un problema de GPU del emulador; validar Impeller en dispositivo real (S12). Host: cmdline-tools 1.1.16717390 y NDK 28.2.13676358 instalados con `android sdk install` (Gradle no pudo instalarlo solo).
- Pendiente: iOS en pausa por decisión del usuario (sin Mac); el proyecto iOS existe, pero no se ha compilado ni probado. Dev contra API http local: S09 debe añadir network_security_config solo en debug, o usar https local. freezed/json_serializable instalados sin uso aún.
- Verificación: `flutter analyze` 0 issues; `flutter test` 8/8 (refresh concurrente único, refresh rechazado, login sin bearer, mapper, arranque en es).

### S07 — Usuarios, historial y endurecimiento — 2026-11-04 — ✅
- Hecho: CRUD `/usuarios`, `GET /clientes/{id}/consumos`, `gosec` v2.30.0 en CI. Desactivar, cambiar rol o resetear contraseña revoca refresh tokens; `ULTIMO_ADMIN` protege con bloqueo al último admin activo.
- Decisiones: `Refresh` bloquea usuario y luego token (evita deadlock con revocación). Username `A-Za-z0-9._-`. `#nosec` con motivo en 4 falsos positivos de gosec.
- Checklist OK: A01/API1/API5, A03, A04, A07, A09, A10, API4 (body 64 KB, size ≤ 50, pool acotado); argon2id y EdDSA de A02; A05 y A06 en backend/imagen (distroless, nonroot, read_only, errores RFC7807, govulncheck, gosec, trivy, Dependabot).
- Checklist pendiente S15: TLS/HSTS, Caddyfile, compose prod, CORS/cabeceras en proxy, `cpus`, firma de imágenes (A08). Móvil MASVS: S08+.
- ADR-006b: cada request valida `activo` y toma el rol de la BD; desactivar o degradar aplica de inmediato. Resetear contraseña no invalida el access token vigente (máx. 15 min).
- Deuda: `WriteTimeout` 15 s vs regla 10 s (margen sobre timeout de handler 10 s). GO-2026-5932 en `x/crypto/openpgp`: no se importa ni se llama; sin versión estable con fix. Rate limit en memoria. `make` no está en este host: comandos ejecutados con `docker run`.
- Verificación: tests verdes; cobertura domain 98.4%, service 93.4%; lint 0; gosec 0 (exit 0); govulncheck 0 llamadas; trivy 0 CRITICAL/HIGH. `-race` en tests nuevos. EXPLAIN historial: usa `consumos_cliente_fecha_idx`, 0.3 ms con 40k filas. E2E en Compose aislado OK, 0 secretos en logs.

### S06 — Consumo — 2026-11-04 — ✅
- Hecho: `POST /consumos` (`ConsumoService`, `ConsumoRepository`), UPDATE condicional atómico, diagnóstico y registro de rechazos, idempotencia con repetición de la respuesta original, rate limit 60/min por usuario.
- Decisiones: los rechazos no guardan la clave y se reevalúan al reintentar. Clave de otro usuario: 409 `IDEMPOTENCY_KEY_CONFLICT`. QR reemplazado: `CLIENTE_NO_EXISTE`. Hasta 3 intentos si una recarga cambia la tiquetera durante el débito. `GetClienteByIDForUpdate` pasa a `FOR NO KEY UPDATE` para evitar deadlock con el FK de `consumos`.
- Pendiente: rate limit en memoria (igual que login): revisar si se escala a varias instancias. Sin cambios en el 501 de historial (S07).
- Verificación: lint 0 issues; tests verdes; cobertura domain 97.8%, service 92.0%; `-race` en http/service/domain y en `TestConsumo_*`. En BD real: 20 goroutines sobre saldo 5 (5 éxitos, saldo 0), borde Bogotá, misma clave concurrente, recarga concurrente sin deadlock. p95 14 ms con `-race` en Docker local (se reporta, no se asevera). E2E en Compose aislado: débito, repetición, SIN_SALDO, QR_INVALIDO, 400 sin clave, 0 tokens en logs. El 429 solo se probó en tests HTTP.

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
