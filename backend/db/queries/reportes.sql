-- name: CountRechazosBetween :one
SELECT count(*) FROM consumos_rechazados
WHERE fecha_hora >= sqlc.arg(desde) AND fecha_hora < sqlc.arg(hasta);

-- name: ConsumosPorDia :many
SELECT (fecha_hora AT TIME ZONE 'America/Bogota')::date AS fecha, count(*) AS total
FROM consumos
WHERE fecha_hora >= sqlc.arg(desde) AND fecha_hora < sqlc.arg(hasta)
  AND (sqlc.narg(cajero_id)::uuid IS NULL OR cajero_id = sqlc.narg(cajero_id))
GROUP BY 1
ORDER BY 1;

-- name: ConsumosPorCajero :many
SELECT c.cajero_id, u.nombre, count(*) AS total
FROM consumos c
JOIN usuarios u ON u.id = c.cajero_id
WHERE c.fecha_hora >= sqlc.arg(desde) AND c.fecha_hora < sqlc.arg(hasta)
  AND (sqlc.narg(cajero_id)::uuid IS NULL OR c.cajero_id = sqlc.narg(cajero_id))
GROUP BY c.cajero_id, u.nombre
ORDER BY total DESC, u.nombre;

-- name: ResumenVigentes :one
SELECT count(*) AS clientes,
       COALESCE(sum(t.saldo), 0)::bigint AS saldo_pendiente,
       count(*) FILTER (WHERE t.saldo > 0 AND t.fecha_vencimiento <= sqlc.arg(limite_vence)::date) AS por_vencer
FROM tiqueteras t
JOIN clientes c ON c.id = t.cliente_id AND c.activo
WHERE t.activa AND t.fecha_vencimiento >= sqlc.arg(hoy)::date;

-- name: VendidosDesde :one
SELECT COALESCE(sum(almuerzos_vendidos), 0)::bigint FROM tiqueteras
WHERE created_at >= sqlc.arg(desde);

-- name: ListSinRecarga :many
SELECT c.id AS cliente_id, c.nombre, c.telefono, t.created_at AS ultima_recarga, t.saldo
FROM clientes c
JOIN tiqueteras t ON t.cliente_id = c.id AND t.activa
WHERE c.activo AND t.created_at < sqlc.arg(corte)
ORDER BY t.created_at, c.id
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CountSinRecarga :one
SELECT count(*) FROM clientes c
JOIN tiqueteras t ON t.cliente_id = c.id AND t.activa
WHERE c.activo AND t.created_at < sqlc.arg(corte);

-- name: ListPorVencer :many
SELECT c.id AS cliente_id, c.nombre, c.telefono, t.fecha_vencimiento, t.saldo
FROM tiqueteras t
JOIN clientes c ON c.id = t.cliente_id AND c.activo
WHERE t.activa AND t.saldo > 0
  AND t.fecha_vencimiento BETWEEN sqlc.arg(hoy)::date AND sqlc.arg(limite)::date
ORDER BY t.fecha_vencimiento, c.id
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CountPorVencer :one
SELECT count(*) FROM tiqueteras t
JOIN clientes c ON c.id = t.cliente_id AND c.activo
WHERE t.activa AND t.saldo > 0
  AND t.fecha_vencimiento BETWEEN sqlc.arg(hoy)::date AND sqlc.arg(limite)::date;

-- name: ListSaldoBajo :many
SELECT c.id AS cliente_id, c.nombre, c.telefono, t.fecha_vencimiento, t.saldo
FROM tiqueteras t
JOIN clientes c ON c.id = t.cliente_id AND c.activo
WHERE t.activa AND t.fecha_vencimiento >= sqlc.arg(hoy)::date AND t.saldo <= sqlc.arg(max_saldo)::int
ORDER BY t.saldo, t.fecha_vencimiento, c.id
LIMIT LEAST(sqlc.arg(size)::int, 50) OFFSET sqlc.arg(page_offset)::int;

-- name: CountSaldoBajo :one
SELECT count(*) FROM tiqueteras t
JOIN clientes c ON c.id = t.cliente_id AND c.activo
WHERE t.activa AND t.fecha_vencimiento >= sqlc.arg(hoy)::date AND t.saldo <= sqlc.arg(max_saldo)::int;
