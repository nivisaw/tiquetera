package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

func TestBogotaDayStart_IsFiveAMUTC(t *testing.T) {
	assert.Equal(t, time.Date(2026, 11, 4, 5, 0, 0, 0, time.UTC), BogotaDayStart(date("2026-11-04")))
}

func TestDashboardDays_SevenDaysEndingToday(t *testing.T) {
	days := DashboardDays(date("2026-11-04"))
	require.Len(t, days, 7)
	assert.Equal(t, date("2026-10-29"), days[0])
	assert.Equal(t, date("2026-11-04"), days[6])
}

func TestFillDays_ZeroForMissing(t *testing.T) {
	days := []time.Time{date("2026-11-01"), date("2026-11-02")}
	got := FillDays(days, []DiaTotal{{Fecha: date("2026-11-02"), Total: 4}})
	assert.Equal(t, []DiaTotal{{Fecha: date("2026-11-01")}, {Fecha: date("2026-11-02"), Total: 4}}, got)
}

func TestDaysBetween_Inclusive(t *testing.T) {
	assert.Len(t, DaysBetween(date("2026-10-30"), date("2026-11-02")), 4)
	assert.Len(t, DaysBetween(date("2026-11-02"), date("2026-11-02")), 1)
}

func TestMonthStart_FirstDay(t *testing.T) {
	assert.Equal(t, date("2026-11-01"), MonthStart(date("2026-11-30")))
}

func TestValidateRango_Bounds(t *testing.T) {
	tests := []struct {
		desde, hasta string
		ok           bool
	}{
		{"2026-11-01", "2026-11-01", true},
		{"2026-10-02", "2026-11-01", true},
		{"2026-10-01", "2026-11-01", false},
		{"2026-11-02", "2026-11-01", false},
	}
	for _, tt := range tests {
		t.Run(tt.desde+"_"+tt.hasta, func(t *testing.T) {
			err := ValidateRango(date(tt.desde), date(tt.hasta))
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			var v ValidationError
			assert.ErrorAs(t, err, &v)
		})
	}
}
