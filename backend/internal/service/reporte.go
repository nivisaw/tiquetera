package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"tiquetera/internal/domain"
	"tiquetera/internal/platform"
)

// ReporteRepository is the persistence port for admin reports. Instants are
// half-open ranges [desde, hasta); dates are Bogotá dates in midnight UTC form.
type ReporteRepository interface {
	CountRechazos(ctx context.Context, desde, hasta time.Time) (int64, error)
	ConsumosPorDia(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) ([]domain.DiaTotal, error)
	ConsumosPorCajero(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) ([]domain.CajeroTotal, error)
	ResumenVigentes(ctx context.Context, hoy, limiteVence time.Time) (domain.ResumenVigentes, error)
	VendidosDesde(ctx context.Context, desde time.Time) (int64, error)
	SinRecarga(ctx context.Context, corte time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error)
	PorVencer(ctx context.Context, hoy, limite time.Time, limit, offset int32) ([]domain.ClienteReporte, int64, error)
	SaldoBajo(ctx context.Context, hoy time.Time, maxSaldo, limit, offset int32) ([]domain.ClienteReporte, int64, error)
}

// ReporteService builds admin reports.
type ReporteService struct {
	repo  ReporteRepository
	clock platform.Clock
}

// NewReporteService returns a ReporteService.
func NewReporteService(repo ReporteRepository, clock platform.Clock) *ReporteService {
	return &ReporteService{repo: repo, clock: clock}
}

func (s *ReporteService) today() time.Time { return domain.TodayInBogota(s.clock.Now()) }

// Dashboard returns the headline figures for today in Bogotá.
func (s *ReporteService) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	hoy := s.today()
	inicioHoy, finHoy := domain.BogotaDayStart(hoy), domain.BogotaDayStart(hoy.AddDate(0, 0, 1))
	days := domain.DashboardDays(hoy)
	var d domain.Dashboard
	porDia, err := s.repo.ConsumosPorDia(ctx, domain.BogotaDayStart(days[0]), finHoy, nil)
	if err != nil {
		return d, fmt.Errorf("consumptions per day: %w", err)
	}
	d.Consumos7d = domain.FillDays(days, porDia)
	d.ConsumosHoy = d.Consumos7d[len(days)-1].Total
	if d.RechazosHoy, err = s.repo.CountRechazos(ctx, inicioHoy, finHoy); err != nil {
		return d, fmt.Errorf("counting rejections: %w", err)
	}
	limite := hoy.AddDate(0, 0, domain.PorVencerDefaultDias)
	if d.Vigentes, err = s.repo.ResumenVigentes(ctx, hoy, limite); err != nil {
		return d, fmt.Errorf("active summary: %w", err)
	}
	if d.VendidosMes, err = s.repo.VendidosDesde(ctx, domain.BogotaDayStart(domain.MonthStart(hoy))); err != nil {
		return d, fmt.Errorf("sold this month: %w", err)
	}
	return d, nil
}

// SinRecarga lists active clients whose last recharge was dias or more days ago.
func (s *ReporteService) SinRecarga(ctx context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error) {
	corte := domain.BogotaDayStart(s.today().AddDate(0, 0, int(-dias)+1))
	items, total, err := s.repo.SinRecarga(ctx, corte, size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("listing without recharge: %w", err)
	}
	return items, total, nil
}

// PorVencer lists active ticket books with balance expiring within dias days.
func (s *ReporteService) PorVencer(ctx context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error) {
	hoy := s.today()
	items, total, err := s.repo.PorVencer(ctx, hoy, hoy.AddDate(0, 0, int(dias)), size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("listing expiring: %w", err)
	}
	return items, total, nil
}

// SaldoBajo lists current ticket books with balance at most maxSaldo.
func (s *ReporteService) SaldoBajo(ctx context.Context, maxSaldo, page, size int32) ([]domain.ClienteReporte, int64, error) {
	items, total, err := s.repo.SaldoBajo(ctx, s.today(), maxSaldo, size, (page-1)*size)
	if err != nil {
		return nil, 0, fmt.Errorf("listing low balance: %w", err)
	}
	return items, total, nil
}

// Consumos summarizes consumptions between two Bogotá dates, inclusive.
func (s *ReporteService) Consumos(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) (domain.ConsumosReporte, error) {
	var r domain.ConsumosReporte
	if err := domain.ValidateRango(desde, hasta); err != nil {
		return r, err
	}
	inicio, fin := domain.BogotaDayStart(desde), domain.BogotaDayStart(hasta.AddDate(0, 0, 1))
	porDia, err := s.repo.ConsumosPorDia(ctx, inicio, fin, cajero)
	if err != nil {
		return r, fmt.Errorf("consumptions per day: %w", err)
	}
	if r.Cajeros, err = s.repo.ConsumosPorCajero(ctx, inicio, fin, cajero); err != nil {
		return r, fmt.Errorf("consumptions per cashier: %w", err)
	}
	r.Dias = domain.FillDays(domain.DaysBetween(desde, hasta), porDia)
	for _, d := range r.Dias {
		r.Total += d.Total
	}
	return r, nil
}
