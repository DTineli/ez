package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/orders"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/go-chi/chi/v5"
)

type AdminOrderHandler struct {
	orderStore   orders.Repository
	contactStore store.ContactStore
	productStore store.ProductStore
}

func NewAdminOrderHandler(
	orderStore orders.Repository,
	contactStore store.ContactStore,
	productStore store.ProductStore,
) *AdminOrderHandler {
	return &AdminOrderHandler{
		orderStore:   orderStore,
		contactStore: contactStore,
		productStore: productStore,
	}
}

func (h *AdminOrderHandler) GetOrdersPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	const perPage = 20
	pageNum := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		pageNum = p
	}

	filters := store.OrderFilters{
		Page:        pageNum,
		PerPage:     perPage,
		ContactName: strings.TrimSpace(r.URL.Query().Get("contact")),
		Status:      store.OrderStatus(r.URL.Query().Get("status")),
	}

	rows, count, err := h.orderStore.ListByTenantPaged(sess.TenantID, filters)
	if err != nil {
		http.Error(w, "Erro ao buscar pedidos", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(count) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	listPage := store.AdminOrderListPage{
		Orders:     rows,
		Filters:    filters,
		TotalPages: totalPages,
		Total:      count,
	}

	if err := Render(templates.AdminOrdersPage(listPage), r, w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *AdminOrderHandler) GetOrderPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	order, err := h.orderStore.GetByID(uint(id), sess.TenantID)
	if err != nil {
		http.Error(w, "Pedido não encontrado", http.StatusNotFound)
		return
	}

	if err := Render(templates.AdminOrderDetailPage(order), r, w); err != nil {
		http.Error(
			w,
			"Error rendering template",
			http.StatusInternalServerError,
		)
	}
}

func (h *AdminOrderHandler) GetNewOrderPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	contacts, err := h.contactStore.FindAll(sess.TenantID, store.ContactFilters{
		Pagination:  store.Pagination{Page: 1, PerPage: 1000},
		ContactType: string(store.Customer),
	})
	if err != nil {
		http.Error(w, "Erro ao buscar contatos", http.StatusInternalServerError)
		return
	}

	if err := Render(templates.AdminNewOrderPage(contacts.Results), r, w); err != nil {
		http.Error(
			w,
			"Error rendering template",
			http.StatusInternalServerError,
		)
	}
}

func (h *AdminOrderHandler) SearchProductsForOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		w.WriteHeader(http.StatusOK)
		return
	}

	results, err := h.productStore.AdminFindAllByUserWithFilters(
		sess.TenantID,
		store.ProductFilters{
			Page:    1,
			PerPage: 10,
			Name:    q,
		},
	)
	if err != nil {
		http.Error(w, "Erro ao buscar produtos", http.StatusInternalServerError)
		return
	}

	// Se não achou por nome, tenta por SKU exato
	if len(results.Results) == 0 {
		results, err = h.productStore.AdminFindAllByUserWithFilters(
			sess.TenantID,
			store.ProductFilters{
				Page:    1,
				PerPage: 10,
				SKU:     q,
			},
		)
		if err != nil {
			http.Error(
				w,
				"Erro ao buscar produtos",
				http.StatusInternalServerError,
			)
			return
		}
	}

	templates.ProductSearchResults(results.Results).Render(r.Context(), w)
}

func (h *AdminOrderHandler) PostNewOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	contactIDStr := r.FormValue("contact_id")
	contactID, err := strconv.ParseUint(contactIDStr, 10, 64)
	if err != nil || contactID == 0 {
		http.Error(w, "Contato inválido", http.StatusBadRequest)
		return
	}

	productIDs := r.Form["product_id[]"]
	variantIDs := r.Form["variant_id[]"]

	quantities := r.Form["quantity[]"]
	unitPrices := r.Form["unit_price[]"]

	if len(productIDs) == 0 || len(productIDs) != len(quantities) ||
		len(productIDs) != len(unitPrices) {
		http.Error(w, "Itens inválidos", http.StatusBadRequest)
		return
	}

	items := make([]store.NewOrderItem, 0, len(productIDs))
	for i := range productIDs {
		pid, err := strconv.ParseUint(productIDs[i], 10, 64)
		if err != nil || pid == 0 {
			continue
		}

		varID, err := strconv.ParseUint(variantIDs[i], 10, 64)
		if err != nil || varID == 0 {
			continue
		}

		qty, err := strconv.Atoi(quantities[i])
		if err != nil || qty <= 0 {
			continue
		}
		price, err := strconv.ParseFloat(unitPrices[i], 64)
		if err != nil || price < 0 {
			continue
		}
		items = append(items, store.NewOrderItem{
			ProductID: uint(pid),
			VariantID: uint(varID),
			Quantity:  qty,
			UnitPrice: price,
		})
	}

	if len(items) == 0 {
		http.Error(w, "Nenhum item válido", http.StatusBadRequest)
		return
	}

	order, err := h.orderStore.Create(sess.TenantID, uint(contactID), items)
	if err != nil {
		http.Error(w, "Erro ao criar pedido", http.StatusInternalServerError)
		return
	}

	HXLocation(w, fmt.Sprintf("/admin/pedidos/%d", order.ID))
}
