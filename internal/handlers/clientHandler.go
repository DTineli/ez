package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
	"gorm.io/gorm"
)

type ClientHandler struct {
	productStore    store.ProductStore
	cartStore       store.CartStore
	orderStore      store.OrderStore
	sessionStore    store.SessionStore
	priceTableStore store.PriceTableStore
}

func NewClientHandler(
	pStore store.ProductStore,
	cStore store.CartStore,
	oStore store.OrderStore,
	sStore store.SessionStore,
	ptStore store.PriceTableStore,
) *ClientHandler {
	return &ClientHandler{
		productStore:    pStore,
		cartStore:       cStore,
		orderStore:      oStore,
		sessionStore:    sStore,
		priceTableStore: ptStore,
	}
}

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

func (c *ClientHandler) GetItemsPage(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)
	isHX := r.Header.Get("HX-Request") == "true"
	const perPage = 9

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	query := r.URL.Query().Get("q")

	products, err := c.productStore.FindAllByUserWithFilters(
		sess.TenantID,
		store.ProductFilters{
			Page:    page,
			PerPage: perPage,
			Search:  query,
		},
	)
	if err != nil {
		ShowToast(w, "Erro ao buscar produtos", "error")
		return
	}

	priceTable, err := c.priceTableStore.GetOne(
		sess.ContactInfo.PriceTable,
		sess.TenantID,
	)
	if err != nil {
		http.Error(
			w,
			"Tabela de preço não encontrada. Contate o administrador.",
			http.StatusUnprocessableEntity,
		)
		return
	}

	var cards []store.CardData
	for _, p := range products.Results {

		variants := make([]store.VariantData, 0, len(p.Variants))
		for _, v := range p.Variants {
			vPrice := v.CostPrice * (1 + priceTable.Percentage/100)

			attrs := make([]store.AttrData, 0, len(v.Attributes))
			for _, a := range v.Attributes {
				attrs = append(attrs, store.AttrData{
					Name:  a.AttributeValue.Attribute.Name,
					Value: a.AttributeValue.Value,
				})
			}

			variants = append(variants, store.VariantData{
				ID:    v.ID,
				Price: vPrice,
				Attrs: attrs,
			})
		}

		cards = append(cards, store.CardData{
			ID:       p.ID,
			Name:     p.Name,
			Variants: variants,
		})
	}

	totalPages := int(math.Ceil(float64(products.Count) / float64(perPage)))
	nextPage := 0
	if page < totalPages {
		nextPage = page + 1
	}

	// HTMX: scroll infinito (page > 1) → só o chunk; busca (page == 1) → conteúdo do grid
	if isHX {
		_ = templates.ClientProductsChunk(cards, nextPage, query).
			Render(r.Context(), w)
		return
	}

	cartCount := int64(0)
	if sess.CartID != 0 {
		if total, err := c.cartStore.CountItems(sess.CartID); err == nil {
			cartCount = total
		}
	}

	RenderClientWithLayout(
		templates.ClientProductsPage(cards, nextPage, query),
		w,
		r,
		cartCount,
		"produtos",
	)
}

func (c *ClientHandler) GetCheckoutPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := middleware.GetSessionFromContext(r)

	cartCount := int64(0)
	items := []store.CartCheckoutItem{}
	totalAmount := 0.0

	var openCart *store.Cart
	var err error

	if sess.CartID != 0 {
		openCart, err = c.cartStore.FindOpenByID(
			sess.CartID,
			sess.TenantID,
			sess.ContactInfo.ID,
		)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ShowToast(w, "Erro ao carregar carrinho", "error")
			return
		}
	}

	if openCart == nil {
		openCart, err = c.cartStore.FindOpenByContact(
			sess.TenantID,
			sess.ContactInfo.ID,
		)
		if err == nil {
			_ = c.sessionStore.SetCartID(r, w, openCart.ID)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			ShowToast(w, "Erro ao carregar carrinho", "error")
			return
		}
	}

	if openCart != nil {
		if total, err := c.cartStore.CountItems(openCart.ID); err == nil {
			cartCount = total
		}

		items, err = c.cartStore.ListCheckoutItems(openCart.ID, sess.TenantID)
		if err != nil {
			ShowToast(w, "Erro ao carregar itens", "error")
			return
		}

		for _, item := range items {
			totalAmount += item.Subtotal
		}
	}

	RenderClientWithLayout(
		templates.ClientCheckoutPage(items, totalAmount),
		w,
		r,
		cartCount,
		"carrinho",
	)
}
