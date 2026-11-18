package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"tiquetera/internal/domain"
	"tiquetera/internal/repository/postgres/sqlcgen"
)

// ReporteRepository implements service.ReporteRepository.
type ReporteRepository struct {
	q *sqlcgen.Queries
}

// NewReporteRepository returns a pool-backed ReporteRepository.
func NewReporteRepository(pool *pgxpool.Pool) *ReporteRepository {
	return &ReporteRepository{q: sqlcgen.New(pool)}
}

// CountRechazos counts rejections in [desde, hasta).
func (r *ReporteRepository) CountRechazos(ctx context.Context, desde, hasta time.Time) (int64, error) {
	n, err := r.q.CountRechazosBetween(ctx, sqlcgen.CountRechazosBetweenParams{Desde: fromTime(desde), Hasta: fromTime(hasta)})
	if err != nil {
		return 0, fmt.Errorf("counting rejections: %w", err)
	}
	return n, nil
}

// ConsumosPorDia counts consumptions per Bogotá date in [desde, hasta).
func (r *ReporteRepository) ConsumosPorDia(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) ([]domain.DiaTotal, error) {
	rows, err := r.q.ConsumosPorDia(ctx, sqlcgen.ConsumosPorDiaParams{Desde: fromTime(desde), Hasta: fromTime(hasta), CajeroID: cajero})
	if err != nil {
		return nil, fmt.Errorf("consumptions per day: %w", err)
	}
	out := make([]domain.DiaTotal, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.DiaTotal(row))
	}
	return out, nil
}

// ConsumosPorCajero counts consumptions per cashier in [desde, hasta).
func (r *ReporteRepository) ConsumosPorCajero(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) ([]domain.CajeroTotal, error) {
	rows, err := r.q.ConsumosPorCajero(ctx, sqlcgen.ConsumosPorCajeroParams{Desde: fromTime(desde), Hasta: fromTime(hasta), CajeroID: cajero})
	if err != nil {
		return nil, fmt.Errorf("consumptions per cashier: %w", err)
	}
	out := make([]domain.CajeroTotal, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.CajeroTotal(row))
	}
	return out, nil
}

// ResumenVigentes summarizes current ticket books.
func (r *ReporteRepository) ResumenVigentes(ctx context.Context, hoy, limiteVence time.Time) (domain.ResumenVigentes, error) {
	row, err := r.q.ResumenVigentes(ctx, sqlcgen.ResumenVigentesParams{Hoy: hoy, LimiteVence: limiteVence})
	if err != nil {
		return domain.ResumenVigentes{}, fmt.Errorf("active summary: %w", err)
	}
	return domain.ResumenVigentes(row), nil
}

// VendidosDesde sums lunches sold in ticket books created since desde.
func (r *ReporteRepository) VendidosDesde(ctx context.Context, desde time.Time) (int64, error) {
	n, err := r.q.VendidosDesde(ctx, fromTime(desde))
	if err != nil {
		return 0, fmt.Errorf("sold since: %w", err)
	}
	return n, nil
}

// SinRecarga lists active clients whose active ticket book predates corte.
func (r *ReporteRepository) SinRecarga(ctx context.Context, corte time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	total, err := r.q.CountSinRecarga(ctx, fromTime(corte))
	if err != nil {
		return nil, 0, fmt.Errorf("counting without recharge: %w", err)
	}
	rows, err := r.q.ListSinRecarga(ctx, sqlcgen.ListSinRecargaParams{Corte: fromTime(corte), Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("listing without recharge: %w", err)
	}
	out := make([]domain.ClienteReporte, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ClienteReporte{
			ClienteID: row.ClienteID,
			Nombre:    row.Nombre,
			Telefono:  row.Telefono,
			Fecha:     row.UltimaRecarga.Time,
			Saldo:     row.Saldo,
		})
	}
	return out, total, nil
}

// PorVencer lists ticket books with balance expiring in [hoy, limite].
func (r *ReporteRepository) PorVencer(ctx context.Context, hoy, limite time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	total, err := r.q.CountPorVencer(ctx, sqlcgen.CountPorVencerParams{Hoy: hoy, Limite: limite})
	if err != nil {
		return nil, 0, fmt.Errorf("counting expiring: %w", err)
	}
	rows, err := r.q.ListPorVencer(ctx, sqlcgen.ListPorVencerParams{Hoy: hoy, Limite: limite, Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("listing expiring: %w", err)
	}
	out := make([]domain.ClienteReporte, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ClienteReporte{
			ClienteID: row.ClienteID,
			Nombre:    row.Nombre,
			Telefono:  row.Telefono,
			Fecha:     row.FechaVencimiento,
			Saldo:     row.Saldo,
		})
	}
	return out, total, nil
}

// SaldoBajo lists current ticket books with balance at most maxSaldo.
func (r *ReporteRepository) SaldoBajo(ctx context.Context, hoy time.Time, maxSaldo, limit, offset int32) ([]domain.ClienteReporte, int64, error) {
	total, err := r.q.CountSaldoBajo(ctx, sqlcgen.CountSaldoBajoParams{Hoy: hoy, MaxSaldo: maxSaldo})
	if err != nil {
		return nil, 0, fmt.Errorf("counting low balance: %w", err)
	}
	rows, err := r.q.ListSaldoBajo(ctx, sqlcgen.ListSaldoBajoParams{Hoy: hoy, MaxSaldo: maxSaldo, Size: limit, PageOffset: offset})
	if err != nil {
		return nil, 0, fmt.Errorf("listing low balance: %w", err)
	}
	out := make([]domain.ClienteReporte, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ClienteReporte{
			ClienteID: row.ClienteID,
			Nombre:    row.Nombre,
			Telefono:  row.Telefono,
			Fecha:     row.FechaVencimiento,
			Saldo:     row.Saldo,
		})
	}
	return out, total, nil
}
