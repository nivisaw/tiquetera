package domain

import (
	"time"

	"github.com/google/uuid"
)

// Report parameter bounds (docs/API.md#reportes-admin).
const (
	SinRecargaDefaultDias = 30
	SinRecargaMaxDias     = 365
	PorVencerDefaultDias  = 7
	PorVencerMaxDias      = 60
	SaldoBajoDefault      = 3
	SaldoBajoMax          = 50
	ConsumosMaxRangoDias  = 31
	dashboardDias         = 7
	day                   = 24 * time.Hour
)

// DiaTotal is a count for one Bogotá calendar date.
type DiaTotal struct {
	Fecha time.Time
	Total int64
}

// CajeroTotal is a cashier's consumption count.
type CajeroTotal struct {
	CajeroID uuid.UUID
	Nombre   string
	Total    int64
}

// ResumenVigentes summarizes current ticket books: clients holding one, the
// lunches still owed and how many with balance expire soon.
type ResumenVigentes struct {
	Clientes       int64
	SaldoPendiente int64
	PorVencer      int64
}

// Dashboard holds the admin's headline figures.
type Dashboard struct {
	ConsumosHoy int64
	Consumos7d  []DiaTotal
	Vigentes    ResumenVigentes
	VendidosMes int64
	RechazosHoy int64
}

// ClienteReporte is a client row in a report; Fecha is the report's date
// (last recharge or expiry).
type ClienteReporte struct {
	ClienteID uuid.UUID
	Nombre    string
	Telefono  string
	Fecha     time.Time
	Saldo     int32
}

// ConsumosReporte summarizes consumptions in a date range.
type ConsumosReporte struct {
	Total   int64
	Dias    []DiaTotal
	Cajeros []CajeroTotal
}

// BogotaDayStart returns the instant a Bogotá date (midnight UTC form) begins.
func BogotaDayStart(date time.Time) time.Time {
	return date.Add(-time.Duration(bogotaOffsetSeconds) * time.Second)
}

// DashboardDays returns the dates of the dashboard trend, oldest first.
func DashboardDays(today time.Time) []time.Time {
	days := make([]time.Time, dashboardDias)
	for i := range days {
		days[i] = today.Add(-time.Duration(dashboardDias-1-i) * day)
	}
	return days
}

// FillDays returns one DiaTotal per date in days, zero where counts lacks it.
func FillDays(days []time.Time, counts []DiaTotal) []DiaTotal {
	byDate := make(map[time.Time]int64, len(counts))
	for _, c := range counts {
		byDate[c.Fecha] = c.Total
	}
	out := make([]DiaTotal, len(days))
	for i, d := range days {
		out[i] = DiaTotal{Fecha: d, Total: byDate[d]}
	}
	return out
}

// DaysBetween lists the dates from desde to hasta inclusive.
func DaysBetween(desde, hasta time.Time) []time.Time {
	var days []time.Time
	for d := desde; !d.After(hasta); d = d.Add(day) {
		days = append(days, d)
	}
	return days
}

// MonthStart returns the first date of the month of a Bogotá date.
func MonthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// ValidateRango checks a consumption report range.
func ValidateRango(desde, hasta time.Time) error {
	if hasta.Before(desde) {
		return ValidationError{"La fecha final no puede ser anterior a la inicial"}
	}
	if hasta.Sub(desde) >= ConsumosMaxRangoDias*day {
		return ValidationError{"El rango no puede superar 31 días"}
	}
	return nil
}
