# Modelo de Datos (PostgreSQL)

Todas las PK `uuid` (`gen_random_uuid()`), tiempos `timestamptz` UTC, `created_at/updated_at` en todas las tablas.

## usuarios
| col | tipo | notas |
|---|---|---|
| id | uuid PK | |
| username | citext UNIQUE | |
| password_hash | text | argon2id |
| nombre | text | |
| rol | text CHECK (rol IN ('admin','cajero')) | |
| activo | bool default true | |
| failed_logins / locked_until | int / timestamptz | bloqueo por fuerza bruta |

## refresh_tokens
id, usuario_id FK, token_hash UNIQUE, family_id, expires_at, revoked_at.

## clientes
| col | tipo | notas |
|---|---|---|
| id | uuid PK | |
| nombre | text NOT NULL (2..120) | |
| telefono | text NOT NULL UNIQUE | E.164 (`+57...`) |
| documento | text NULL | opcional |
| email | citext NULL | |
| qr_token_hash | bytea UNIQUE NOT NULL | SHA-256 |
| qr_version | int default 1 | se incrementa al regenerar |
| activo | bool default true | |
| created_by | uuid FK usuarios | |

## tiqueteras
Una tiquetera **activa** por cliente (índice parcial `UNIQUE (cliente_id) WHERE activa`).
| col | tipo | notas |
|---|---|---|
| id | uuid PK | |
| cliente_id | uuid FK | |
| almuerzos_comprados | int CHECK (> 0 AND <= 500) | |
| saldo | int CHECK (saldo >= 0) | |
| fecha_vencimiento | date NOT NULL | |
| activa | bool | al crear una nueva, la anterior se desactiva |
| created_by | uuid FK usuarios | |

## consumos (append-only, nunca UPDATE/DELETE)
| col | tipo |
|---|---|
| id | uuid PK |
| tiquetera_id | uuid FK |
| cliente_id | uuid FK |
| cajero_id | uuid FK usuarios |
| fecha_hora | timestamptz default now() |
| saldo_resultante | int |
| idempotency_key | uuid UNIQUE |

Índices: `consumos(cliente_id, fecha_hora DESC)`, `consumos(cajero_id, fecha_hora DESC)`.

## consumos_rechazados (auditoría)
id, cajero_id, cliente_id NULL, motivo text, fecha_hora. Para detectar fraude/abuso.

## audit_log
id, usuario_id, accion, entidad, entidad_id, datos jsonb (sin secretos), fecha_hora.
