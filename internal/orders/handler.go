package orders

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	ordertemplates "github.com/DTineli/ez/internal/orders/templates"
	"github.com/DTineli/ez/internal/store"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) PostGeneratePickListPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}
	idStrs := r.Form["ids"]

	sess := middleware.GetSessionFromContext(r)

	var ids []uint
	for _, s := range idStrs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			ids = append(ids, uint(id))
		}
	}

	orders, err := h.service.FetchOrderInfo(ids, sess.TenantID)

	if err != nil {
		http.Error(w, "Erro ao gerar Lista", http.StatusBadRequest)
		return
	}

	ordertemplates.PickList(orders).Render(r.Context(), w)
}

func (h *Handler) PostBulkStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	idStrs := r.Form["ids"]
	status := store.OrderStatus(r.FormValue("status"))
	if status == "" || len(idStrs) == 0 {
		http.Error(w, "IDs e status obrigatórios", http.StatusBadRequest)
		return
	}

	sess := middleware.GetSessionFromContext(r)

	var ids []uint
	for _, s := range idStrs {
		id, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			ids = append(ids, uint(id))
		}
	}

	h.service.BulkAtualizarStatus(
		ids,
		sess.TenantID,
		status,
		store.OrderAtorSeller,
	)

	w.Header().Set("HX-Redirect", "/admin/pedidos/")
	w.WriteHeader(http.StatusOK)
}

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

	status := store.OrderStatus(r.FormValue("status"))
	if status == "" {
		http.Error(w, "Status obrigatório", http.StatusBadRequest)
		return
	}

	sess := middleware.GetSessionFromContext(r)

	if err := h.service.AtualizarStatus(uint(id), sess.TenantID, status, store.OrderAtorSeller); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/admin/pedidos/%d", id))
	w.WriteHeader(http.StatusOK)
}
