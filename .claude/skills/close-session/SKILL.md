---
name: close-session
description: Cerrar una sesión del plan. Usar al terminar el alcance de una sesión o antes de agotar contexto.
---

1. Ejecuta la verificación de la sesión (`make lint test` y/o `flutter analyze && flutter test`), mostrando solo el resumen.
2. Revisa el diff contra `.claude/rules/security.md` (secretos, logs con PII, SQL concatenado, RBAC).
3. Si cambió el contrato, sincroniza `docs/API.md` y `openapi.yaml`.
4. Agrega entrada en `docs/PROGRESS.md` usando la plantilla (≤ 8 líneas) y actualiza "Siguiente sesión".
5. Si quedó incompleta: marca ⚠️ y lista exactamente qué falta para que otro agente continúe.
6. Mensaje final al usuario: ≤ 5 líneas (estado, verificación, siguiente paso).
