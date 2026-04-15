package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
)

type ClientHandler struct {
	productStore store.ProductStore
}

func NewClientHandler(
	pStore store.ProductStore,
) *ClientHandler {
	return &ClientHandler{
		productStore: pStore,
	}
}

func RenderClient(c templ.Component, w http.ResponseWriter, r *http.Request) error {
	if r.Header.Get("HX-Request") == "true" {
		return c.Render(r.Context(), w)
	}

	return templates.
		Layout_Client(c).
		Render(r.Context(), w)
}

func (c *ClientHandler) GetItemsPage(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)

	products, err := c.productStore.FindAllByUser(sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao buscar pedidos", "error")
	}

	var cards []store.CardData

	for _, p := range products {
		cards = append(cards, store.CardData{
			ID:         p.ID,
			Name:       p.Name,
			Price:      p.CostPrice,
			Photo_Link: "",
		})
	}

	RenderClient(templates.ClientProductsPage(cards), w, r)
}
