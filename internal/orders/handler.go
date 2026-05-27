package orders

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PatchStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	status := Status(r.FormValue("status"))
	if status == "" {
		http.Error(w, "Status obrigatório", http.StatusBadRequest)
		return
	}

	sess := middleware.GetSessionFromContext(r)

	if err := h.service.AtualizarStatus(uint(id), sess.TenantID, status, AtorSeller); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/admin/pedidos/%d", id))
	w.WriteHeader(http.StatusOK)
}
