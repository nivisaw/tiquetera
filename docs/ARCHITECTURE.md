# Arquitectura

```
[Flutter App] --HTTPS/JSON--> [Caddy :443] --> [API Go :8080] --> [PostgreSQL :5432]
   admin / cajero               TLS, headers     red interna Docker    volumen persistente
```

## Backend — arquitectura hexagonal ligera
```
backend/
  cmd/api/main.go            wiring, config, graceful shutdown
  internal/
    config/                  env → struct validado (fail fast)
    domain/                  entidades, errores de dominio, reglas puras (sin deps externas)
    service/                 casos de uso (interfaces de repos inyectadas)
    repository/postgres/     implementación sqlc
    http/                    handlers, middleware (auth, rbac, ratelimit, requestid, recover), DTOs
    auth/                    hashing argon2id, JWT, refresh
    platform/                logger, clock (inyectable para tests), idgen
  db/migrations/             golang-migrate (up/down)
  db/queries/                SQL para sqlc
```
Regla de dependencias: `http → service → domain ← repository`. `domain` no importa nada del proyecto.

## Móvil — feature-first + capas
```
mobile/lib/
  core/        config (flavors dev/prod), http (dio + interceptores), errores, theme, router
  features/
    auth/      data/ domain/ presentation/
    clients/   (admin) alta, detalle, paquetes, QR, compartir
    scanner/   (cajero) cámara, resultado
  shared/      widgets reutilizables
```
Estado: Riverpod (`AsyncNotifier`). Navegación por rol con guards de go_router.

## Flujo crítico: consumo
1. Cajero escanea → app extrae token (valida prefijo `TQ1.`) → `POST /v1/consumos` con `Idempotency-Key` (UUID por escaneo).
2. API: hash token → en **una transacción**:
   `UPDATE tiqueteras SET saldo = saldo - 1 WHERE cliente_id = $1 AND activa AND saldo > 0 AND fecha_vencimiento >= $hoy RETURNING saldo`
   Si 0 filas → diagnostica motivo (no existe / vencida / sin saldo / inactiva). Si 1 fila → `INSERT consumos`.
3. Respuesta < 300 ms p95. App muestra pantalla verde/roja.
4. Anti doble-escaneo: idempotency key + bloqueo de re-escaneo del mismo token en la app por 5 s.

## Escalabilidad
- API stateless (JWT) → escalable horizontalmente detrás de Caddy.
- Pool pgx dimensionado por env. Índices en `token_hash`, `consumos(cliente_id, fecha_hora)`.
- Multi-sede futura: columna `sede_id` prevista en ADR futuro, no implementar aún.

## Observabilidad
- Logs JSON (slog) con `request_id`, sin PII ni tokens. `/healthz` (liveness) y `/readyz` (BD).
