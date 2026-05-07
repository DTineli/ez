package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/DTineli/ez/internal/templates/components"
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

func (c *ClientHandler) RenderSelectTableByClient(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := middleware.GetSessionFromContext(r)
	tables, err := c.priceTableStore.FindAllActiveByTenantAndClient(
		sess.TenantID,
		sess.ContactInfo.ID,
	)
	if err != nil {
		ShowToast(w, "Erro ao recuperar dados", "error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	options := make([]components.SelectOption, 0, len(tables))
	for _, table := range tables {
		options = append(options, components.SelectOption{
			Value: strconv.Itoa(int(table.ID)),
			Label: table.Name,
		})
	}

	Render(components.Select(components.SelectParams{
		Placeholder: "Selecione uma tabela",
		Label:       "Tabela de Preço",
		Name:        "price_table",
		Selected:    r.URL.Query().Get("selected"),
		Options:     options,
	}), r, w)
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
		5,
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
				ID:        v.ID,
				Price:     vPrice,
				IsDefault: v.IsDefault,
				Attrs:     attrs,
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

	if isHX {
		_ = templates.ClientProductsChunk(cards, nextPage, query).
			Render(r.Context(), w)
		return
	}

	RenderClientWithLayout(
		templates.ClientProductsPage(cards, nextPage, query),
		w,
		r,
		c.getCartCount(sess),
		"produtos",
	)
}

func (c *ClientHandler) GetCheckoutPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := middleware.GetSessionFromContext(r)

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
		c.getCartCount(sess),
		"carrinho",
	)
}
