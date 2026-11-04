package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"tiquetera/internal/domain"
)

// UsuarioService is the account-management use-case port.
type UsuarioService interface {
	Create(ctx context.Context, actor uuid.UUID, input domain.NewUser) (domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)
	List(ctx context.Context, page, size int32) ([]domain.User, int64, error)
	Update(ctx context.Context, actor, id uuid.UUID, patch domain.UserPatch) (domain.User, error)
	ResetPassword(ctx context.Context, actor, id uuid.UUID, password string) error
}

type createUsuarioRequest struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Nombre   string      `json:"nombre"`
	Rol      domain.Role `json:"rol"`
}

type updateUsuarioRequest struct {
	Nombre optional[string]      `json:"nombre"`
	Rol    optional[domain.Role] `json:"rol"`
	Activo optional[bool]        `json:"activo"`
}

// toPatch rejects empty updates and nulls: every field is required when present.
func (r updateUsuarioRequest) toPatch() (domain.UserPatch, bool) {
	if !r.Nombre.Set && !r.Rol.Set && !r.Activo.Set {
		return domain.UserPatch{}, false
	}
	if nullOn(r.Nombre) || nullOn(r.Rol) || nullOn(r.Activo) {
		return domain.UserPatch{}, false
	}
	return domain.UserPatch{Nombre: r.Nombre.Value, Rol: r.Rol.Value, Activo: r.Activo.Value}, true
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

type usuarioDTO struct {
	ID       uuid.UUID   `json:"id"`
	Username string      `json:"username"`
	Nombre   string      `json:"nombre"`
	Rol      domain.Role `json:"rol"`
	Activo   bool        `json:"activo"`
}

type usuarioPage struct {
	Items []usuarioDTO `json:"items"`
	Page  int32        `json:"page"`
	Size  int32        `json:"size"`
	Total int64        `json:"total"`
}

func toUsuarioDTO(u domain.User) usuarioDTO {
	return usuarioDTO{ID: u.ID, Username: u.Username, Nombre: u.Nombre, Rol: u.Rol, Activo: u.Activo}
}

type usuariosHandler struct {
	log *slog.Logger
	svc UsuarioService
}

func (h *usuariosHandler) create(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorFrom(w, r)
	if !ok {
		return
	}
	var req createUsuarioRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	created, err := h.svc.Create(r.Context(), actor, domain.NewUser{
		Username: req.Username, Password: req.Password, Nombre: req.Nombre, Rol: req.Rol,
	})
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toUsuarioDTO(created))
}

func (h *usuariosHandler) list(w http.ResponseWriter, r *http.Request) {
	page, size, ok := parsePagination(w, r)
	if !ok {
		return
	}
	users, total, err := h.svc.List(r.Context(), page, size)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	items := make([]usuarioDTO, 0, len(users))
	for _, u := range users {
		items = append(items, toUsuarioDTO(u))
	}
	writeJSON(w, http.StatusOK, usuarioPage{Items: items, Page: page, Size: size, Total: total})
}

func (h *usuariosHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := clienteID(w, r)
	if !ok {
		return
	}
	user, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(h.log, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toUsuarioDTO(user))
}

func (h *usuariosHandler) update(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := actorAndCliente(w, r)
	if !ok {
		return
	}
	var req updateUsuarioRequest
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
	writeJSON(w, http.StatusOK, toUsuarioDTO(updated))
}

func (h *usuariosHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	actor, id, ok := actorAndCliente(w, r)
	if !ok {
		return
	}
	var req resetPasswordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := h.svc.ResetPassword(r.Context(), actor, id, req.Password); err != nil {
		writeError(h.log, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
