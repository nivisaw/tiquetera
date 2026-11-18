# Plan de Desarrollo por Sesiones

Cada sesión es independiente: un agente nuevo lee `CLAUDE.md` → `docs/PROGRESS.md` → su sesión aquí. Duración objetivo: 1 contexto de agente. No avanzar sin cumplir DoD (CLAUDE.md).

Formato: **Objetivo · Entradas · Tareas · Entregables · Verificación**

---
## Fase 0 — Fundaciones

### S01 Monorepo e infraestructura local
- **Entradas:** ADR, rules/docker.md, rules/git.md
- **Tareas:** `git init`, `.gitignore`, `.editorconfig`, `Makefile` raíz; `infra/docker-compose.yml` (db postgres:16-alpine con healthcheck y volumen; api placeholder), `infra/.env.example`; `.github/workflows/ci.yml` esqueleto.
- **Entregables:** `make up` levanta Postgres sano.
- **Verificación:** `docker compose ps` healthy; no hay secretos commiteados.

### S02 Esqueleto backend Go
- **Tareas:** `go mod init`, estructura de ARCHITECTURE.md, config por env validada, slog JSON, chi con middlewares (request-id, recover, timeout, body limit 64KB, security headers), `/healthz` `/readyz`, graceful shutdown, Dockerfile multi-stage distroless nonroot, golangci-lint config, `make test/lint`.
- **Verificación:** `make up` → `curl /healthz` 200; lint y tests en verde.

### S03 Migraciones, sqlc y OpenAPI
- **Tareas:** migraciones de todas las tablas de DATA_MODEL.md (up/down), `sqlc.yaml`, queries base, `backend/api/openapi.yaml` desde API.md, seed de admin vía comando `cmd/seed` (contraseña por env, nunca hardcode).
- **Verificación:** `make migrate` up/down/up idempotente; test de integración con testcontainers verifica constraints (saldo<0 falla, tiquetera activa única).

## Fase 1 — Backend de dominio

### S04 Autenticación y RBAC
- **Tareas:** argon2id, JWT EdDSA (clave por env/archivo), login/refresh/logout, rotación con detección de reutilización, bloqueo tras 5 intentos (15 min), middleware `RequireRole`, rate limit por IP+usuario, audit_log de login.
- **Verificación:** tests de: credenciales malas, token expirado, refresh reutilizado revoca familia, cajero → 403 en /clientes.

### S05 Clientes, tiqueteras y QR
- **Tareas:** casos de uso crear cliente+tiquetera (transacción), generación token (crypto/rand) y hash, listar/buscar paginado, recarga, regenerar QR, validaciones (E.164, rangos, fecha Bogotá con clock inyectable), audit_log.
- **Verificación:** unit tests de servicio (mocks de repo) + integración; token nunca aparece en logs ni en GET.

### S06 Consumo (núcleo crítico)
- **Tareas:** `POST /consumos` con UPDATE condicional atómico, diagnóstico de motivo, registro de rechazos, idempotencia (UNIQUE + devolver respuesta previa), rate limit por cajero.
- **Verificación:** test de concurrencia: 20 goroutines sobre saldo=5 → exactamente 5 éxitos y saldo 0; vencido ayer rechaza, vence hoy acepta; misma idempotency key no debita dos veces. p95 < 50 ms en BD local.

### S07 Usuarios, historial y endurecimiento
- **Tareas:** CRUD usuarios (admin), historial de consumos paginado, revisión OWASP API Top 10 contra rules/security.md, `govulncheck`, `gosec` en CI.
- **Verificación:** checklist de security.md marcado en PROGRESS.md.

## Fase 2 — App móvil

### S08 Esqueleto Flutter
- **Tareas:** `flutter create` (org `co.<empresa>.tiquetera`), flavors dev/prod (`--dart-define`), estructura feature-first, theme (rules/ux-ui.md), go_router, dio con interceptores (auth, refresh con mutex, errores RFC7807 → tipos de dominio), `very_good_analysis`.
- **Verificación:** `flutter analyze` 0 issues; app arranca en emulador Android y simulador iOS.

### S09 Login y sesión por rol
- **Tareas:** pantalla login, almacenamiento en flutter_secure_storage, refresh automático, logout, redirección por rol (admin → Clientes, cajero → Escáner), manejo sin conexión.
- **Verificación:** widget tests de login (éxito/error/cargando); test de guard por rol.

### S10 Admin: registro de cliente y QR
- **Tareas:** formulario (nombre, teléfono con prefijo +57, documento, email opcional, almuerzos stepper, date picker vencimiento) con validación; al guardar mostrar QR grande (qr_flutter) con nombre y saldo; listado/búsqueda de clientes; detalle con recarga y regenerar QR.
- **Verificación:** widget/golden tests del formulario y de la pantalla QR.

### S11 Compartir por WhatsApp
- **Tareas:** render QR a PNG (RepaintBoundary) en directorio temporal, `share_plus` con texto; botón secundario wa.me con teléfono del cliente; borrar archivo temporal tras compartir.
- **Verificación:** prueba manual documentada en Android e iOS; unit test del builder de URL wa.me (encoding).

### S12 Cajero: escáner y feedback
- **Tareas:** permiso de cámara con explicación y fallback a ajustes, `mobile_scanner` (solo formato QR), debounce 5 s por token, Idempotency-Key por escaneo, pantalla éxito (verde, check animado, saldo en tipografía ≥ 64sp, háptico + sonido) y rechazo (rojo, motivo legible mapeado de `code`), auto-retorno al escáner en 3 s, linterna.
- **Verificación:** widget tests de ambos estados y mapeo de todos los `code`; prueba en dispositivo real.

### S13 Historial y pulido UX
- **Tareas:** historial de consumos por cliente, estados vacíos/errores/carga, accesibilidad (Semantics, contraste, escalado de texto), i18n es-CO con `intl`.
- **Verificación:** checklist de rules/ux-ui.md.

### S13B Admin: dashboard y reportes
- **Tareas:** pestaña "Reportes" del admin (ver `docs/API.md#reportes-admin`). Fechas y "hoy" en Bogotá.
  - **Dashboard:** almuerzos servidos hoy y últimos 7 días (barras por día), clientes con tiquetera vigente, almuerzos prepagados pendientes por servir (suma de saldos vigentes), almuerzos vendidos en el mes (recargas), tiqueteras que vencen en 7 días, rechazos de hoy.
  - **Sin recarga ≥ N días** (default 30): clientes activos cuya última tiquetera se creó hace ≥ N días; muestra fecha de última recarga y saldo.
  - **Por vencer:** tiqueteras vigentes con saldo > 0 que vencen en ≤ N días (default 7), para avisar al cliente antes de que pierda almuerzos.
  - **Saldo bajo:** tiqueteras vigentes con saldo ≤ N (default 3), oportunidad de recarga.
  - **Consumos por día:** rango `desde/hasta` (máx. 31 días), total por día y por cajero (cuadre de caja).
  - Cada fila de cliente abre su detalle; botón "Abrir chat" (wa.me) en sin recarga, por vencer y saldo bajo.
- **Backend:** migración que añade `tiqueteras.almuerzos_vendidos` (lo vendido en esa recarga; hoy `almuerzos_comprados` incluye el saldo arrastrado), con backfill = comprados − saldo de la tiquetera anterior; queries sqlc con `EXPLAIN` revisado e índices si hace falta (nueva migración); listados paginados (size ≤ 50); solo `admin` (tests 403 cajero).
- **Verificación:** tests de integración por reporte con bordes de fecha en Bogotá (día 29/30, vence hoy/mañana); widget tests de dashboard y listados (cargando/vacío/error).

## Fase 3 — Calidad, despliegue y release

### S14 Tests E2E
- **Tareas:** `integration_test` Flutter contra stack docker local: login admin → crear cliente → login cajero → consumir → saldo correcto; casos de rechazo.
- **Verificación:** E2E en verde en CI (emulador Android).

### S15 Despliegue VPS
- **Tareas:** `infra/docker-compose.prod.yml` (caddy, api, db sin puertos públicos, límites de recursos, restart policies, read_only, no-new-privileges), Caddyfile con TLS/HSTS, backups `pg_dump` diarios cifrados con retención 14 días + script de restauración probado, guía `docs/DEPLOY.md` (firewall ufw 22/80/443, usuario no-root, SSH por llave).
- **Verificación:** restauración de backup probada; SSL Labs A.

### S16 Release a tiendas
- **Tareas:** iconos/splash, firmado Android (keystore fuera del repo), configuración iOS (permisos `NSCameraUsageDescription`), ofuscación (`--obfuscate --split-debug-info`), política de privacidad, `docs/RELEASE.md`, opcional fastlane.
- **Verificación:** build AAB y IPA generados; checklist de tiendas.

## Backlog (no implementar sin aprobación)
Modo offline del cajero · notificación de saldo bajo/vencimiento.
