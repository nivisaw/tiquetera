# Progreso (bitácora de relevo entre agentes)

Regla: entradas breves (≤ 8 líneas por sesión). Lo más reciente arriba. Solo hechos útiles para el siguiente agente.

## Estado actual
- **Siguiente sesión:** S02
- **Bloqueos:** ninguno

## Historial
<!-- Plantilla:
### SXX — <título> — AAAA-MM-DD — ✅/⚠️
- Hecho:
- Decisiones/desvíos:
- Pendiente/deuda:
- Comandos de verificación:
-->

### S01 — Monorepo e infraestructura local — 2026-11-03 — ✅
- Hecho: git (rama `feat/s01-infra`), .gitignore/.editorconfig/.gitattributes, Makefile (up/down/ps/logs/clean), infra/docker-compose.yml (db postgres:16.6-alpine healthy; api en profile `api`), infra/.env.example, CI (validación compose + gitleaks).
- Decisiones: Postgres expuesto solo en `127.0.0.1:5442` (5432/5433 ocupados por otros proyectos en el host).
- Pendiente: `make` no está instalado en Windows del host → usar Git Bash con make, o `docker compose --env-file infra/.env -f infra/docker-compose.yml ...`. S02 debe añadir jobs backend al CI y quitar el profile del api.
- Verificación: `docker compose --env-file infra/.env -f infra/docker-compose.yml ps` → db healthy.

### S00 — Documentación inicial — 2026-11-03 — ✅
- Hecho: CLAUDE.md, docs/, .claude/rules, .claude/skills.
