# Seguridad (OWASP Top 10 2021 + API Security Top 10 2023 + MASVS)

## Checklist obligatorio
| Riesgo | Control |
|---|---|
| A01 Broken Access Control / API1 BOLA / API5 BFLA | RBAC en servidor en cada ruta (`RequireRole`); nunca confiar en el rol que envía la app. IDs uuid; validar pertenencia/permiso en servicio. Tests 403 por rol. |
| A02 Cryptographic Failures | TLS 1.2+ (Caddy), HSTS. argon2id (m=64MB,t=3,p=2). Token QR 32 bytes `crypto/rand`; guardar solo SHA-256. JWT EdDSA. Nada sensible en SharedPreferences: usar flutter_secure_storage. |
| A03 Injection | Solo SQL parametrizado (sqlc). Validar y normalizar toda entrada (longitudes, E.164, rangos). Prohibido concatenar SQL. |
| A04 Insecure Design | Débito atómico, idempotencia, rate limit, auditoría de rechazos, límites de negocio (almuerzos ≤ 500). |
| A05 Security Misconfiguration | Contenedores nonroot, read_only, distroless; BD sin puerto público; CORS cerrado (app móvil no lo necesita); cabeceras seguras; errores genéricos RFC7807. Debug desactivado en prod. |
| A06 Vulnerable Components | Versiones fijadas (go.sum, pubspec.lock, imágenes por digest/tag exacto). CI: `govulncheck`, `gosec`, `trivy` sobre imagen, Dependabot. |
| A07 Auth Failures | Bloqueo tras 5 fallos, rate limit login, refresh rotativo con detección de reutilización, access 15 min, logout revoca. Contraseñas ≥ 10 caracteres. |
| A08 Integrity Failures | CI firma/etiqueta imágenes; no ejecutar scripts remotos (`curl | sh`) en Dockerfiles. Migraciones versionadas. |
| A09 Logging Failures | audit_log de acciones admin y login; logs JSON sin PII, contraseñas, tokens ni payload QR. |
| A10 SSRF | El backend no hace requests a URLs proporcionadas por usuarios. |
| API4 Resource Consumption | Body ≤ 64KB, paginación size ≤ 50, timeouts (read 10s, write 10s), pool BD acotado. |

## Móvil (MASVS)
- Release con `--obfuscate`; sin logs de datos sensibles (`kReleaseMode`).
- Android: `usesCleartextTraffic=false`, `allowBackup=false`. iOS: ATS activo.
- Certificate pinning: opcional (decidir en S16, riesgo de bloqueo al renovar certificados).
- Screenshot del QR permitido (necesario para compartir); no mostrar token en texto.

## Secretos
- Solo por variables de entorno / archivos montados (`*_FILE`). `.env` en `.gitignore`; mantener `.env.example` sin valores reales.
- Si detectas un secreto commiteado: detente y avisa al usuario.
