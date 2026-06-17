package handlers

import (
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/services"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func RenderClientWithLayout(
	c templ.Component,
	w http.ResponseWriter,
	r *http.Request,
	cartCount int64,
	activeTab string,
) error {
	if r.Header.Get("HX-Request") == "true" {
		return c.Render(r.Context(), w)
	}

	return templates.
		Layout_Client(c, cartCount, activeTab).
		Render(r.Context(), w)
}

func (c *ClientHandler) getCartCount(sess *store.Session) int64 {
	if sess.CartID == 0 {
		return 0
	}
	total, err := c.cartStore.CountItems(sess.CartID)
	if err != nil {
		return 0
	}
	return total
}

func applyPrice(table store.PriceTable, variant store.Variant) float64 {
	return services.ApplyPriceTable(variant.CostPrice, &table)
}

func applyCheckoutPrice(costPrice float64, table *store.PriceTable) float64 {
	return services.ApplyPriceTable(costPrice, table)
}

func queryParamUintOrZero(r *http.Request, paramName string) uint64 {
	val, err := strconv.ParseUint(r.URL.Query().Get(paramName), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func parseQueryParamUint(
	w http.ResponseWriter,
	r *http.Request,
	paramName, emptyMsg string,
) (uint64, bool) {
	val, err := strconv.ParseUint(r.URL.Query().Get(paramName), 10, 64)
	if err != nil || val == 0 {
		ShowToast(w, emptyMsg, "error")
		return 0, false
	}
	return val, true
}

func parseURLParamUint(
	w http.ResponseWriter,
	r *http.Request,
	paramName, errorMsg string,
) (uint64, bool) {
	val, err := strconv.ParseUint(chi.URLParam(r, paramName), 10, 64)
	if err != nil || val == 0 {
		ShowToast(w, errorMsg, "error")
		return 0, false
	}
	return val, true
}
