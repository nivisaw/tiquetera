# Reglas de Testing

## Pirámide y cobertura
- Backend: unit (servicio/dominio, table-driven) + integración (repositorios y HTTP con testcontainers Postgres). Cobertura ≥ 80% en `domain` y `service`.
- Móvil: unit (notifiers, repositorios con mocktail) + widget + golden (pantallas QR, éxito, rechazo) + `integration_test` E2E (S14). Cobertura ≥ 70% en `features/*/domain|data`.

## Casos obligatorios
- Consumo: éxito, cada código de rechazo, borde de vencimiento (hoy acepta / ayer rechaza, en zona Bogotá), concurrencia (no saldo negativo), idempotencia.
- Auth: 401/403 por rol, refresh reutilizado, bloqueo de cuenta.
- Validación de entradas inválidas en cada endpoint.

## Prácticas
- Tests deterministas: reloj inyectado, sin `sleep`, sin red externa.
- Nombre: `Test<Unidad>_<Escenario>_<Esperado>`.
- Al corregir un bug, primero un test que lo reproduzca.
- Ejecutar solo tests afectados durante el desarrollo; suite completa antes de cerrar sesión.
