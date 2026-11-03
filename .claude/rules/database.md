# Reglas de Base de Datos

- Todo cambio de esquema = nueva migración `NNNN_descripcion.up.sql` + `.down.sql`. Nunca editar migraciones ya aplicadas.
- Restricciones en BD además de en código: `CHECK`, `NOT NULL`, `UNIQUE`, FKs con `ON DELETE RESTRICT`.
- `consumos` es append-only. Clientes se desactivan, no se borran.
- Toda query nueva en `db/queries/*.sql` → `make sqlc`. Revisar `EXPLAIN` en queries de listado/consumo.
- Paginación por keyset o `LIMIT/OFFSET` con tope 50.
- Usuario de BD de la app sin privilegios de superusuario; migraciones con usuario owner separado.
- Backups: ver S15.
