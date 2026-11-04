# Contrato API v1

Base: `/v1`. JSON. Auth: `Authorization: Bearer <access>`. Fechas ISO-8601. Fuente de verdad: `backend/api/openapi.yaml` (se crea en S03 y debe mantenerse sincronizado con este archivo).

## Errores (RFC 7807)
```json
{ "type": "about:blank", "title": "...", "status": 422, "code": "TIQUETERA_VENCIDA", "detail": "..." }
```
Nunca exponer stack traces ni SQL.

## Auth
| Método | Ruta | Rol | Notas |
|---|---|---|---|
| POST | /auth/login | público | `{username,password}` → `{access_token, refresh_token, expires_in, user{id,nombre,rol}}`. Rate limit 5/min/IP+user (429 `RATE_LIMITED`). Bloqueo 15 min tras 5 fallos. Credenciales malas, inactivo y bloqueado: mismo 401 `INVALID_CREDENTIALS` |
| POST | /auth/refresh | público | `{refresh_token}`; rota refresh. Inválido, vencido o reutilizado: 401 `INVALID_REFRESH_TOKEN`; reutilizado revoca familia |
| POST | /auth/logout | auth | `{refresh_token}`; revoca familia. 204 también si el token es desconocido o ajeno (no revoca nada) |

Sin token válido: 401 `UNAUTHENTICATED`. Rol insuficiente: 403 `FORBIDDEN`. Entrada inválida: 400 `INVALID_REQUEST`.

## Clientes (admin)
| Método | Ruta | Notas |
|---|---|---|
| POST | /clientes | `{nombre, telefono, documento?, email?, almuerzos, fecha_vencimiento}` → `201 {cliente, tiquetera, qr_payload}`. **qr_payload solo se devuelve aquí y en regenerar** |
| GET | /clientes?q=&page=&size= | búsqueda por nombre/teléfono, paginado (size ≤ 50) |
| GET | /clientes/{id} | incluye tiquetera activa |
| PATCH | /clientes/{id} | datos básicos |
| POST | /clientes/{id}/tiqueteras | nueva recarga `{almuerzos, fecha_vencimiento}`; desactiva la anterior |
| POST | /clientes/{id}/qr/regenerar | invalida QR anterior → `{qr_payload}` |
| GET | /clientes/{id}/consumos | historial paginado |

Listados: `{items, page, size, total}` (`page` desde 1).
Validaciones: teléfono E.164 (default país +57), almuerzos 1..500, fecha_vencimiento >= hoy (Bogotá).
Errores: JSON mal formado, campo desconocido o null en campo obligatorio de PATCH 400 `INVALID_REQUEST`; valor inválido 422 `VALIDATION_ERROR`; teléfono repetido 409 `TELEFONO_DUPLICADO`; recarga a cliente inactivo 422 `CLIENTE_INACTIVO`; id inexistente o mal formado 404 `NOT_FOUND`.
PATCH: campo ausente no cambia; `documento`/`email` en null se borran. La recarga suma el saldo anterior, aunque la tiquetera previa haya vencido: la nueva queda con `almuerzos_comprados` = `saldo` = recarga + saldo anterior; si supera 500, 422 `VALIDATION_ERROR` y no cambia nada.
Historial: más reciente primero; `items[]` con `{id, tiquetera_id, cliente_id, cajero_id, fecha_hora, saldo_resultante}`.

## Consumos (cajero, admin)
`POST /consumos` · Header `Idempotency-Key: <uuid>` (obligatorio) · Body `{ "qr": "TQ1.xxxxx" }`

`200`:
```json
{ "ok": true, "cliente": {"nombre": "Ana P."}, "saldo_restante": 7, "fecha_vencimiento": "2026-12-31", "fecha_hora": "..." }
```
`422` con `code`: `QR_INVALIDO` · `CLIENTE_NO_EXISTE` · `CLIENTE_INACTIVO` · `SIN_TIQUETERA` · `TIQUETERA_VENCIDA` · `SIN_SALDO`
Repetir misma Idempotency-Key → misma respuesta original, sin nuevo débito. Un rechazo no debita ni guarda la clave: reintentar lo evalúa de nuevo. Clave usada por otro usuario: 409 `IDEMPOTENCY_KEY_CONFLICT`. Header ausente o inválido, o `qr` vacío o > 128 bytes: 400 `INVALID_REQUEST`.
QR mal formado: `QR_INVALIDO`. QR bien formado desconocido o reemplazado al regenerar: `CLIENTE_NO_EXISTE`.
Rate limit: 60/min por cajero.

## Usuarios (admin)
| Método | Ruta | Notas |
|---|---|---|
| POST | /usuarios | `{username, password, nombre, rol}` → `201 {id, username, nombre, rol, activo}` |
| GET | /usuarios?page=&size= | paginado (size ≤ 50), orden por username |
| GET | /usuarios/{id} | |
| PATCH | /usuarios/{id} | `{nombre?, rol?, activo?}`; no se borran usuarios, se desactivan |
| PUT | /usuarios/{id}/password | `{password}` → 204; limpia el bloqueo |

Validaciones: username 3..50 (`A-Za-z0-9._-`, único sin distinguir mayúsculas), contraseña 10..128 caracteres, nombre 2..120, rol `admin`/`cajero`.
Errores: JSON mal formado, campo desconocido, PATCH vacío o con null 400 `INVALID_REQUEST`; valor inválido 422 `VALIDATION_ERROR`; username repetido 409 `USERNAME_DUPLICADO`; desactivar o degradar al único admin activo 422 `ULTIMO_ADMIN`; id inexistente o mal formado 404 `NOT_FOUND`.
Desactivar, cambiar de rol o restablecer la contraseña revoca los refresh tokens del usuario. Desactivar o cambiar de rol aplica de inmediato también a los access tokens ya emitidos (ADR-006b); restablecer la contraseña no invalida el access token vigente (máx. 15 min).

## Reportes (admin, fase 2)
`GET /reportes/consumos?desde=&hasta=&cajero_id=`

## Salud
`GET /healthz` · `GET /readyz` (sin auth, sin detalles internos)
