package http

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"tiquetera/internal/domain"
)

// ReporteService is the admin-report use-case port.
type ReporteService interface {
	Dashboard(ctx context.Context) (domain.Dashboard, error)
	SinRecarga(ctx context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error)
	PorVencer(ctx context.Context, dias, page, size int32) ([]domain.ClienteReporte, int64, error)
	SaldoBajo(ctx context.Context, maxSaldo, page, size int32) ([]domain.ClienteReporte, int64, error)
	Consumos(ctx context.Context, desde, hasta time.Time, cajero *uuid.UUID) (domain.ConsumosReporte, error)
}

type diaTotalDTO struct {
	Fecha string `json:"fecha"`
	Total int64  `json:"total"`
}

type dashboardDTO struct {
	ConsumosHoy      int64         `json:"consumos_hoy"`
	Consumos7d       []diaTotalDTO `json:"consumos_7d"`
	ClientesVigentes int64         `json:"clientes_vigentes"`
	SaldoPendiente   int64         `json:"saldo_pendiente"`
	VendidosMes      int64         `json:"vendidos_mes"`
	PorVencer7d      int64         `json:"por_vencer_7d"`
	RechazosHoy      int64         `json:"rechazos_hoy"`
}

type sinRecargaDTO struct {
	ClienteID     uuid.UUID `json:"cliente_id"`
	Nombre        string    `json:"nombre"`
	Telefono      string    `json:"telefono"`
	UltimaRecarga time.Time `json:"ultima_recarga"`
	Saldo         int32     `json:"saldo"`
}

type vencimientoDTO struct {
	ClienteID        uuid.UUID `json:"cliente_id"`
	Nombre           string    `json:"nombre"`
	Telefono         string    `json:"telefono"`
	FechaVencimiento string    `json:"fecha_vencimiento"`
	Saldo            int32     `json:"saldo"`
}

type reportePage[T any] struct {
	Items []T   `json:"items"`
	Page  int32 `json:"page"`
	Size  int32 `json:"size"`
	Total int64 `json:"total"`
}

type cajeroTotalDTO struct {
	CajeroID uuid.UUID `json:"cajero_id"`
	Nombre   string    `json:"nombre"`
	Total    int64     `json:"total"`
}

type consumosReporteDTO struct {
	Total   int64            `json:"total"`
	Dias    []diaTotalDTO    `json:"dias"`
	Cajeros []cajeroTotalDTO `json:"cajeros"`
}

type reportesHandler struct {
	log *slog.Logger
	svc ReporteService
}

func (h *reportesHandler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.Dashboard(r.Context())
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, dashboardDTO{
		ConsumosHoy:      d.ConsumosHoy,
		Consumos7d:       toDiaTotalDTOs(d.Consumos7d),
		ClientesVigentes: d.Vigentes.Clientes,
		SaldoPendiente:   d.Vigentes.SaldoPendiente,
		VendidosMes:      d.VendidosMes,
		PorVencer7d:      d.Vigentes.PorVencer,
		RechazosHoy:      d.RechazosHoy,
	})
}

func (h *reportesHandler) sinRecarga(w http.ResponseWriter, r *http.Request) {
	dias, ok := h.queryInt(w, r, "dias", domain.SinRecargaDefaultDias, 1, domain.SinRecargaMaxDias)
	if !ok {
		return
	}
	h.clientList(w, r, dias, h.svc.SinRecarga, func(c domain.ClienteReporte) any {
		return sinRecargaDTO{ClienteID: c.ClienteID, Nombre: c.Nombre, Telefono: c.Telefono, UltimaRecarga: c.Fecha, Saldo: c.Saldo}
	})
}

func (h *reportesHandler) porVencer(w http.ResponseWriter, r *http.Request) {
	dias, ok := h.queryInt(w, r, "dias", domain.PorVencerDefaultDias, 0, domain.PorVencerMaxDias)
	if !ok {
		return
	}
	h.clientList(w, r, dias, h.svc.PorVencer, toVencimientoDTO)
}

func (h *reportesHandler) saldoBajo(w http.ResponseWriter, r *http.Request) {
	maxSaldo, ok := h.queryInt(w, r, "max", domain.SaldoBajoDefault, 0, domain.SaldoBajoMax)
	if !ok {
		return
	}
	h.clientList(w, r, maxSaldo, h.svc.SaldoBajo, toVencimientoDTO)
}

type clientReport func(ctx context.Context, param, page, size int32) ([]domain.ClienteReporte, int64, error)

func (h *reportesHandler) clientList(w http.ResponseWriter, r *http.Request, param int32, fetch clientReport, toDTO func(domain.ClienteReporte) any) {
	page, size, ok := parsePagination(w, r)
	if !ok {
		return
	}
	rows, total, err := fetch(r.Context(), param, page, size)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	items := make([]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDTO(row))
	}
	writeJSON(w, http.StatusOK, reportePage[any]{Items: items, Page: page, Size: size, Total: total})
}

func (h *reportesHandler) consumos(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	desde, errDesde := time.Parse(time.DateOnly, query.Get("desde"))
	hasta, errHasta := time.Parse(time.DateOnly, query.Get("hasta"))
	if errDesde != nil || errHasta != nil {
		writeError(h.log, w, r, domain.ValidationError{Message: "Indica las fechas desde y hasta (AAAA-MM-DD)"})
		return
	}
	var cajero *uuid.UUID
	if raw := query.Get("cajero_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(h.log, w, r, domain.ValidationError{Message: "Cajero inválido"})
			return
		}
		cajero = &id
	}
	rep, err := h.svc.Consumos(r.Context(), desde, hasta, cajero)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	cajeros := make([]cajeroTotalDTO, 0, len(rep.Cajeros))
	for _, c := range rep.Cajeros {
		cajeros = append(cajeros, cajeroTotalDTO(c))
	}
	writeJSON(w, http.StatusOK, consumosReporteDTO{Total: rep.Total, Dias: toDiaTotalDTOs(rep.Dias), Cajeros: cajeros})
}

func toVencimientoDTO(c domain.ClienteReporte) any {
	return vencimientoDTO{
		ClienteID:        c.ClienteID,
		Nombre:           c.Nombre,
		Telefono:         c.Telefono,
		FechaVencimiento: c.Fecha.Format(time.DateOnly),
		Saldo:            c.Saldo,
	}
}

func toDiaTotalDTOs(days []domain.DiaTotal) []diaTotalDTO {
	out := make([]diaTotalDTO, 0, len(days))
	for _, d := range days {
		out = append(out, diaTotalDTO{Fecha: d.Fecha.Format(time.DateOnly), Total: d.Total})
	}
	return out
}

// queryInt reads an optional integer parameter in [lo, hi]; out of range is 422.
func (h *reportesHandler) queryInt(w http.ResponseWriter, r *http.Request, name string, fallback, lo, hi int32) (int32, bool) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n < int64(lo) || n > int64(hi) {
		writeError(h.log, w, r, domain.ValidationError{Message: "Parámetro " + name + " fuera de rango"})
		return 0, false
	}
	return int32(n), true
}
