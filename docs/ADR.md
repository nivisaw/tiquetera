# Decisiones de Arquitectura (ADR)

Formato: Decisión · Motivo · Descartado. Para cambiar una, añade un ADR nuevo con estado "Propuesto".

## ADR-001 Backend en Go
- **Motivo:** binario único, imagen Docker distroless ~15 MB, bajo consumo de RAM en VPS, concurrencia nativa, tipado estricto que reduce errores de agentes IA.
- **Descartado:** Node.js (más dependencias transitivas → mayor superficie de supply-chain, imagen más pesada).

## ADR-002 PostgreSQL 16
- **Motivo:** transacciones y bloqueo de fila robustos (débito atómico), `CHECK` constraints, `timestamptz`, `uuid`, índices parciales.
- **Descartado:** MySQL (CHECK/tipos de fecha menos estrictos).

## ADR-003 Flutter
- **Motivo:** UI idéntica en Android/iOS, rendimiento AOT, `mobile_scanner` (ML Kit/AVFoundation) maduro, una sola base de código y tooling (`flutter analyze`) estricto.
- **Descartado:** React Native (más variabilidad por dependencias nativas de terceros).

## ADR-004 Acceso a datos con sqlc + pgx (sin ORM)
- **Motivo:** SQL explícito, tipado generado, consultas parametrizadas por construcción (anti-inyección).

## ADR-005 QR = token opaco aleatorio
- QR contiene `TQ1.<token>` donde token = 32 bytes aleatorios base64url. En BD se guarda **solo SHA-256(token)**. Nunca IDs secuenciales ni datos personales en el QR. Revocable/regenerable.

## ADR-006 Autenticación
- Login usuario/contraseña (argon2id). Access JWT (EdDSA, 15 min) + refresh token opaco rotativo (30 días, hash en BD, detección de reutilización). Roles en servidor: `admin`, `cajero`.

## ADR-007 Compartir por WhatsApp
- Primario: `share_plus` comparte la imagen PNG del QR + texto (hoja nativa → WhatsApp).
- Secundario: deep link `https://wa.me/<E164>?text=...` (solo texto, no adjunta imagen).
- WhatsApp Cloud API: fuera del MVP (requiere cuenta Business y plantillas aprobadas).

## ADR-007b Zona horaria
- Negocio en `America/Bogota`. Vigencia: `current_date (Bogota) <= fecha_vencimiento` (inclusive). BD guarda `timestamptz` UTC; `fecha_vencimiento` es `date`.

## ADR-008 Reverse proxy Caddy
- TLS automático (Let's Encrypt), HSTS, cabeceras de seguridad. Solo Caddy expone puertos 80/443; BD nunca expuesta.
