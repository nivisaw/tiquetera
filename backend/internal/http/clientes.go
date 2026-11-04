package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tiquetera/internal/domain"
	"tiquetera/internal/service"
)

const (
	defaultPageSize = 20
	maxPageSize     = 50
	maxPage         = 1_000_000
	maxSearchRunes  = 120
)

// ClienteService is the client-management use-case port.
type ClienteService interface {
	Create(ctx context.Context, actor uuid.UUID, data domain.ClienteData, input domain.TiqueteraInput) (service.ClienteCreated, error)
	Get(ctx context.Context, id uuid.UUID) (service.ClienteDetail, error)
	List(ctx context.Context, q string, page, size int32) ([]domain.Cliente, int64, error)
	Update(ctx context.Context, actor, id uuid.UUID, patch domain.ClientePatch) (domain.Cliente, error)
	Recharge(ctx context.Context, actor, id uuid.UUID, input domain.TiqueteraInput) (domain.Tiquetera, error)
	RegenerateQR(ctx context.Context, actor, id uuid.UUID) (string, error)
	ListConsumos(ctx context.Context, id uuid.UUID, page, size int32) ([]domain.ConsumoEntry, int64, error)
}

type createClienteRequest struct {
	Nombre           string  `json:"nombre"`
	Telefono         string  `json:"telefono"`
	Documento        *string `json:"documento"`
	Email            *string `json:"email"`
	Almuerzos        int     `json:"almuerzos"`
	FechaVencimiento string  `json:"fecha_vencimiento"`
}

type createTiqueteraRequest struct {
	Almuerzos        int    `json:"almuerzos"`
	FechaVencimiento string `json:"fecha_vencimiento"`
}

// optional tells an absent JSON field (Set false) from null (Value nil).
type optional[T any] struct {
	Set   bool
	Value *T
}

// UnmarshalJSON records that the field was present.
func (o *optional[T]) UnmarshalJSON(b []byte) error {
	o.Set = true
	return json.Unmarshal(b, &o.Value)
}

type updateClienteRequest struct {
	Nombre    optional[string] `json:"nombre"`
	Telefono  optional[string] `json:"telefono"`
	Documento optional[string] `json:"documento"`
	Email     optional[string] `json:"email"`
	Activo    optional[bool]   `json:"activo"`
}

// toPatch rejects empty updates and nulls on required fields.
func (r updateClienteRequest) toPatch() (domain.ClientePatch, bool) {
	if !r.Nombre.Set && !r.Telefono.Set && !r.Documento.Set && !r.Email.Set && !r.Activo.Set {
		return domain.ClientePatch{}, false
	}
	if nullOn(r.Nombre) || nullOn(r.Telefono) || nullOn(r.Activo) {
		return domain.ClientePatch{}, false
	}
	return domain.ClientePatch{
		Nombre:       r.Nombre.Value,
		Telefono:     r.Telefono.Value,
		SetDocumento: r.Documento.Set,
		Documento:    r.Documento.Value,
		SetEmail:     r.Email.Set,
		Email:        r.Email.Value,
		Activo:       r.Activo.Value,
	}, true
}

func nullOn[T any](o optional[T]) bool { return o.Set && o.Value == nil }

type clienteDTO struct {
	ID        uuid.UUID `json:"id"`
	Nombre    string    `json:"nombre"`
	Telefono  string    `json:"telefono"`
	Documento *string   `json:"documento"`
	Email     *string   `json:"email"`
	Activo    bool      `json:"activo"`
	QRVersion int32     `json:"qr_version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tiqueteraDTO struct {
	ID                 uuid.UUID `json:"id"`
	ClienteID          uuid.UUID `json:"cliente_id"`
	AlmuerzosComprados int32     `json:"almuerzos_comprados"`
	Saldo              int32     `json:"saldo"`
	FechaVencimiento   string    `json:"fecha_vencimiento"`
	Activa             bool      `json:"activa"`
	CreatedAt          time.Time `json:"created_at"`
}

type clienteDetalleDTO struct {
	clienteDTO
	Tiquetera *tiqueteraDTO `json:"tiquetera"`
}

type createClienteResponse struct {
	Cliente   clienteDTO   `json:"cliente"`
	Tiquetera tiqueteraDTO `json:"tiquetera"`
	QRPayload string       `json:"qr_payload"`
}

type qrResponse struct {
	QRPayload string `json:"qr_payload"`
}

type clientePage struct {
	Items []clienteDTO `json:"items"`
	Page  int32        `json:"page"`
	Size  int32        `json:"size"`
	Total int64        `json:"total"`
}

type consumoDTO struct {
	ID              uuid.UUID `json:"id"`
	TiqueteraID     uuid.UUID `json:"tiquetera_id"`
	ClienteID       uuid.UUID `json:"cliente_id"`
	CajeroID        uuid.UUID `json:"cajero_id"`
	FechaHora       time.Time `json:"fecha_hora"`
	SaldoResultante int32     `json:"saldo_resultante"`
}

type consumoPage struct {
	Items []consumoDTO `json:"items"`
	Page  int32        `json:"page"`
	Size  int32        `json:"size"`
	Total int64        `json:"total"`
}

func toClienteDTO(c domain.Cliente) clienteDTO {
	return clienteDTO{
		ID: c.ID, Nombre: c.Nombre, Telefono: c.Telefono, Documento: c.Documento, Email: c.Email,
		Activo: c.Activo, QRVersion: c.QRVersion, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func toTiqueteraDTO(t domain.Tiquetera) tiqueteraDTO {
	return tiqueteraDTO{
		ID: t.ID, ClienteID: t.ClienteID, AlmuerzosComprados: t.AlmuerzosComprados, Saldo: t.Saldo,
		FechaVencimiento: t.FechaVencimiento.Format(time.DateOnly), Activa: t.Activa, CreatedAt: t.CreatedAt,
	}
}

type clientesHandler struct {
	log *slog.Logger
	svc ClienteService
}

func (h *clientesHandler) create(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}
	var req createClienteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	created, err := h.svc.Create(r.Context(), actor,
		domain.ClienteData{Nombre: req.Nombre, Telefono: req.Telefono, Documento: req.Documento, Email: req.Email},
		domain.TiqueteraInput{Almuerzos: req.Almuerzos, FechaVencimiento: req.FechaVencimiento})
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, createClienteResponse{
		Cliente:   toClienteDTO(created.Cliente),
		Tiquetera: toTiqueteraDTO(created.Tiquetera),
		QRPayload: created.QRPayload,
	})
}

func (h *clientesHandler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	q := strings.TrimSpace(query.Get("q"))
	page, okPage := parseBoundedInt(query.Get("page"), 1, maxPage)
	size, okSize := parseBoundedInt(query.Get("size"), defaultPageSize, maxPageSize)
	if !okPage || !okSize || utf8.RuneCountInString(q) > maxSearchRunes || strings.ContainsRune(q, 0) {
		writeInvalidRequest(w)
		return
	}
	clientes, total, err := h.svc.List(r.Context(), q, page, size)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	items := make([]clienteDTO, 0, len(clientes))
	for _, c := range clientes {
		items = append(items, toClienteDTO(c))
	}
	writeJSON(w, http.StatusOK, clientePage{Items: items, Page: page, Size: size, Total: total})
}

func (h *clientesHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := clienteID(w, r)
	if !ok {
		return
	}
	detail, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	resp := clienteDetalleDTO{clienteDTO: toClienteDTO(detail.Cliente)}
	if detail.Tiquetera != nil {
		t := toTiqueteraDTO(*detail.Tiquetera)
		resp.Tiquetera = &t
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *clientesHandler) update(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := actorAndCliente(w, r)
	if !ok {
		return
	}
	var req updateClienteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	patch, valid := req.toPatch()
	if !valid {
		writeInvalidRequest(w)
		return
	}
	updated, err := h.svc.Update(r.Context(), actor, id, patch)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toClienteDTO(updated))
}

func (h *clientesHandler) recharge(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := actorAndCliente(w, r)
	if !ok {
		return
	}
	var req createTiqueteraRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	created, err := h.svc.Recharge(r.Context(), actor, id,
		domain.TiqueteraInput{Almuerzos: req.Almuerzos, FechaVencimiento: req.FechaVencimiento})
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTiqueteraDTO(created))
}

func (h *clientesHandler) regenerateQR(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := actorAndCliente(w, r)
	if !ok {
		return
	}
	payload, err := h.svc.RegenerateQR(r.Context(), actor, id)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, qrResponse{QRPayload: payload})
}

func (h *clientesHandler) listConsumos(w http.ResponseWriter, r *http.Request) {
	id, ok := clienteID(w, r)
	if !ok {
		return
	}
	page, size, ok := parsePagination(w, r)
	if !ok {
		return
	}
	consumos, total, err := h.svc.ListConsumos(r.Context(), id, page, size)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	items := make([]consumoDTO, 0, len(consumos))
	for _, c := range consumos {
		items = append(items, consumoDTO(c))
	}
	writeJSON(w, http.StatusOK, consumoPage{Items: items, Page: page, Size: size, Total: total})
}

func parsePagination(w http.ResponseWriter, r *http.Request) (page, size int32, ok bool) {
	query := r.URL.Query()
	page, okPage := parseBoundedInt(query.Get("page"), 1, maxPage)
	size, okSize := parseBoundedInt(query.Get("size"), defaultPageSize, maxPageSize)
	if !okPage || !okSize {
		writeInvalidRequest(w)
		return 0, 0, false
	}
	return page, size, true
}

func actorFrom(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	principal, ok := principalFrom(r.Context())
	if !ok {
		writeUnauthenticated(w)
		return uuid.Nil, false
	}
	return principal.UserID, true
}

// clienteID answers 404 for a malformed id: no such resource can exist.
func clienteID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeProblem(w, http.StatusNotFound, "NOT_FOUND", "Recurso no encontrado")
		return uuid.Nil, false
	}
	return id, true
}

func actorAndCliente(w http.ResponseWriter, r *http.Request) (actor, id uuid.UUID, ok bool) {
	if actor, ok = actorFrom(w, r); !ok {
		return uuid.Nil, uuid.Nil, false
	}
	if id, ok = clienteID(w, r); !ok {
		return uuid.Nil, uuid.Nil, false
	}
	return actor, id, true
}

// parseBoundedInt reads an optional integer in 1..limit.
func parseBoundedInt(raw string, fallback, limit int32) (int32, bool) {
	if raw == "" {
		return fallback, true
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n < 1 || n > int64(limit) {
		return 0, false
	}
	return int32(n), true
}
