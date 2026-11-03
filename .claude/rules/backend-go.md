# Reglas Backend Go

- Go 1.27, módulos. `gofmt`, `goimports`, `golangci-lint` (errcheck, govet, staticcheck, gosec, revive, bodyclose, sqlclosecheck, contextcheck).
- Respeta capas de docs/ARCHITECTURE.md. Handlers delgados: parsear → validar → servicio → DTO.
- `context.Context` primer parámetro en toda función con I/O. Nunca `context.Background()` dentro de request.
- Errores: envolver con `fmt.Errorf("...: %w", err)`; errores de dominio tipados (`domain.ErrSinSaldo`) mapeados a HTTP en un solo lugar (`http/errors.go`).
- Prohibido `panic` en flujo normal, variables globales mutables, `init()` con lógica.
- Tiempo: usar `platform.Clock` inyectado (testeable); zona `America/Bogota` para reglas de vigencia.
- Transacciones vía helper `WithTx(ctx, fn)`; nada de lógica de negocio en SQL salvo el UPDATE condicional del débito.
- Dependencias: stdlib primero. Aprobadas: chi, pgx/v5, sqlc, golang-migrate, golang-jwt/jwt/v5, x/crypto/argon2, google/uuid, testify, testcontainers-go, httprate. Otra dependencia → justificar en PROGRESS.md.
- Config por env con valores validados al arranque; fallar rápido.
- Funciones ≤ 50 líneas aprox., nombres explícitos en inglés para código; mensajes de usuario en español.
