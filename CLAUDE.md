# CLAUDE.md — Tiquetera de Almuerzos

App móvil (Android/iOS) para gestionar tiqueteras prepagadas de almuerzos con QR. Roles: **admin** (registra clientes, asigna paquetes, genera/comparte QR) y **cajero** (escanea QR, debita 1 almuerzo).

## Stack (decidido — no re-litigar, ver docs/ADR.md)
- **Backend:** Go 1.23+ · chi (router) · pgx/v5 · sqlc · golang-migrate · slog
- **BD:** PostgreSQL 16
- **Móvil:** Flutter 3.x (Dart 3) · Riverpod · go_router · dio · mobile_scanner · qr_flutter · share_plus · flutter_secure_storage
- **Infra:** Docker + Docker Compose (api, db, caddy) en VPS · GitHub Actions CI

## Estructura
```
backend/   API Go (ver .claude/rules/backend-go.md)
mobile/    App Flutter (ver .claude/rules/mobile-flutter.md)
infra/     docker-compose, Caddyfile, scripts
docs/      ARCHITECTURE, DATA_MODEL, API, ADR, PLAN, PROGRESS
```

## Reglas de comportamiento del agente (OBLIGATORIAS — optimización de tokens)
1. **Respuestas mínimas.** Sin saludos, preámbulos, resúmenes repetidos ni explicaciones no pedidas. Reporta solo: qué se hizo, qué falló, qué decide el usuario.
2. **No repitas información** ya presente en el chat o en docs; referencia el archivo (`docs/API.md#consumos`).
3. **Lee solo lo necesario.** Empieza por `docs/PROGRESS.md` y la sesión actual de `docs/PLAN.md`. No explores el repo completo; usa Grep/Glob dirigidos y lee rangos de líneas.
4. **No re-decidas** lo que está en `docs/ADR.md`. Si algo debe cambiar, propón un ADR nuevo en 2-3 líneas y espera aprobación.
5. **Una sesión = un alcance.** No adelantes tareas de otras sesiones ni refactorices fuera del alcance.
6. **Actúa, no preguntes**, cuando haya un default razonable; pregunta solo decisiones de negocio bloqueantes.
7. **Sin código comentado, sin TODOs huérfanos, sin archivos de ejemplo innecesarios.**
8. **Ediciones pequeñas** (Edit) en vez de reescribir archivos completos.
9. Salida de comandos: filtra (`| tail -n 30`, `-q`, `--reporter=compact`) para no inundar el contexto.
10. Al terminar la sesión ejecuta la skill `close-session` (actualiza PROGRESS.md). Al iniciar, `start-session`.

## Definición de terminado (DoD) por sesión
- Compila, `make lint` y `make test` en verde (backend y/o mobile según aplique).
- Tests nuevos para la lógica añadida (ver .claude/rules/testing.md).
- Sin secretos en el repo. Contrato de `docs/API.md` respetado o actualizado.
- `docs/PROGRESS.md` actualizado.

## Comandos (se crean en sesiones S01/S02)
- `make up` / `make down` — levanta stack local con Docker Compose
- `make test` · `make lint` · `make migrate` · `make sqlc`
- `cd mobile && flutter test && flutter analyze`

## Reglas detalladas
`.claude/rules/`: security · backend-go · mobile-flutter · database · testing · docker · ux-ui · git
