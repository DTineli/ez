package handlers

import (
	"net/http"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type AdminOrderHandler struct {
	orderStore store.OrderStore
}

func NewAdminOrderHandler(orderStore store.OrderStore) *AdminOrderHandler {
	return &AdminOrderHandler{orderStore: orderStore}
}

func (h *AdminOrderHandler) GetOrdersPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	orders, err := h.orderStore.ListByTenant(sess.TenantID)
	if err != nil {
		http.Error(w, "Erro ao buscar pedidos", http.StatusInternalServerError)
		return
	}

	if err := Render(templates.AdminOrdersPage(orders), r, w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
