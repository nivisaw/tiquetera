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
| POST | /auth/login | público | `{username,password}` → `{access_token, refresh_token, expires_in, user{id,nombre,rol}}`. Rate limit 5/min/IP+user |
| POST | /auth/refresh | público | rota refresh |
| POST | /auth/logout | auth | `{refresh_token}`; revoca familia |

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

## Consumos (cajero, admin)
`POST /consumos` · Header `Idempotency-Key: <uuid>` (obligatorio) · Body `{ "qr": "TQ1.xxxxx" }`

`200`:
```json
{ "ok": true, "cliente": {"nombre": "Ana P."}, "saldo_restante": 7, "fecha_vencimiento": "2026-12-31", "fecha_hora": "..." }
```
`422` con `code`: `QR_INVALIDO` · `CLIENTE_NO_EXISTE` · `CLIENTE_INACTIVO` · `SIN_TIQUETERA` · `TIQUETERA_VENCIDA` · `SIN_SALDO`
Repetir misma Idempotency-Key → misma respuesta original, sin nuevo débito.
Rate limit: 60/min por cajero.

## Usuarios (admin)
CRUD básico `/usuarios` (crear cajero, desactivar vía `PATCH /usuarios/{id}`, reset con `PUT /usuarios/{id}/password`).

## Reportes (admin, fase 2)
`GET /reportes/consumos?desde=&hasta=&cajero_id=`

## Salud
`GET /healthz` · `GET /readyz` (sin auth, sin detalles internos)
