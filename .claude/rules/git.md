# Reglas Git / CI

- Rama por sesión: `feat/sXX-<slug>`. Commits Conventional Commits (`feat:`, `fix:`, `test:`, `chore:`, `docs:`), pequeños y atómicos.
- Commit/push solo cuando el usuario lo pida.
- CI (GitHub Actions) en cada PR: backend lint + test + govulncheck + gosec; mobile analyze + test; build de imagen + trivy. Bloquea merge si falla.
- Nunca `--no-verify` ni forzar push a `main`.
