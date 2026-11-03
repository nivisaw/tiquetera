---
name: security-check
description: Auditoría rápida OWASP del código cambiado. Usar en S07, antes de releases o cuando se toquen auth, consumos o QR.
---

1. Recorre la tabla de `.claude/rules/security.md`; para cada fila verifica el control en el código (Grep dirigido, no lectura completa).
2. Ejecuta (vía Docker si aplica): `govulncheck ./...`, `gosec ./...`, `trivy image <api>`, `flutter pub outdated`.
3. Reporta solo hallazgos: `severidad · archivo:línea · problema · fix`. Sin hallazgos → "Sin hallazgos".
4. Corrige críticos/altos dentro del alcance; el resto va a PROGRESS.md como deuda.
