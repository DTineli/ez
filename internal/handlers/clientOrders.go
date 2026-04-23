package handlers

import (
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/templates"
	"github.com/go-chi/chi/v5"
)

func (c *ClientHandler) GetOrdersPage(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)

	orders, err := c.orderStore.ListByContact(sess.TenantID, sess.ContactInfo.ID)
	if err != nil {
		ShowToast(w, "Erro ao buscar pedidos", "error")
		return
	}

	cartCount := int64(0)
	if sess.CartID != 0 {
		if total, err := c.cartStore.CountItems(sess.CartID); err == nil {
			cartCount = total
		}
	}

	RenderClientWithLayout(templates.ClientOrdersPage(orders), w, r, cartCount, "pedidos")
}

func (c *ClientHandler) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	order, err := c.orderStore.GetByID(uint(id), sess.TenantID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if order.ContactID != sess.ContactInfo.ID {
		http.NotFound(w, r)
		return
	}

	cartCount := int64(0)
	if sess.CartID != 0 {
		if total, err := c.cartStore.CountItems(sess.CartID); err == nil {
			cartCount = total
		}
	}

	RenderClientWithLayout(templates.ClientOrderDetailPage(order), w, r, cartCount, "pedidos")
}
