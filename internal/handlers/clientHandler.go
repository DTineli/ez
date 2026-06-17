package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/orders"
	"github.com/DTineli/ez/internal/services"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/DTineli/ez/internal/templates/components"
	"gorm.io/gorm"
)

type ClientHandler struct {
	productStore  store.ProductStore
	cartStore     store.CartStore
	orderStore    orders.Repository
	orderService  *orders.Service
	sessionStore  store.SessionStore
	priceTableSvc services.PriceTableService
	contactStore  store.ContactStore
}

func NewClientHandler(
	pStore store.ProductStore,
	cStore store.CartStore,
	oStore orders.Repository,
	sStore store.SessionStore,
	ptSvc services.PriceTableService,
	ccStore store.ContactStore,
) *ClientHandler {
	return &ClientHandler{
		productStore:  pStore,
		cartStore:     cStore,
		orderStore:    oStore,
		orderService:  orders.NewService(oStore),
		sessionStore:  sStore,
		priceTableSvc: ptSvc,
		contactStore:  ccStore,
	}
}

func (c *ClientHandler) RenderCheckoutContent(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := middleware.GetSessionFromContext(r)

	items := []store.CartCheckoutItem{}
	totalAmount := 0.0

	var openCart *store.Cart
	var err error

	// buscar tabela no banco
	price_tableID := queryParamUintOrZero(r, "price_table")

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

		var pt *store.PriceTable
		if price_tableID != 0 {
			fetched, err := c.priceTableSvc.GetOne(
				uint(price_tableID),
				sess.TenantID,
			)
			if err == nil {
				pt = fetched
			}
		}

		for i := range items {
			items[i].UnitPrice = c.priceTableSvc.Apply(items[i].CostPrice, pt)
			items[i].Subtotal = items[i].UnitPrice * float64(items[i].Quantity)
			totalAmount += items[i].Subtotal
		}
	}

	fmt.Println(items)

	showPrice := price_tableID != 0
	Render(templates.ClientCartContent(items, totalAmount, showPrice), r, w)
}

func (c *ClientHandler) RenderSelectTableByClient(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := middleware.GetSessionFromContext(r)

	tables, err := c.contactStore.FindContactPriceTables(
		sess.ContactInfo.ID,
		sess.TenantID,
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
	query := r.URL.Query().Get("q")

	tables, err := c.contactStore.FindContactPriceTables(
		sess.ContactInfo.ID,
		sess.TenantID,
	)
	if err != nil {
		ShowToast(w, "Erro ao recuperar tabelas", "error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	options := make([]components.SelectOption, 0, len(tables))
	for _, t := range tables {
		options = append(options, components.SelectOption{
			Value: strconv.Itoa(int(t.ID)),
			Label: t.Name,
		})
	}

	selectParams := components.SelectParams{
		Placeholder: "Selecione uma tabela",
		Label:       "Tabela de Preço",
		Name:        "price_table",
		Options:     options,
	}

	RenderClientWithLayout(
		templates.ClientProductsPage(query, selectParams),
		w,
		r,
		c.getCartCount(sess),
		"produtos",
	)

}

func (c *ClientHandler) FetchItems(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)
	const perPage = 9

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	query := r.URL.Query().Get("q")
	priceTable, ok := parseQueryParamUint(
		w,
		r,
		"price_table",
		"Selecione uma tabela de preço",
	)
	if !ok {
		return
	}

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

	prices, err := c.priceTableSvc.GetOne(
		uint(priceTable),
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

	cards := makeCardData(products.Results, *prices, c.priceTableSvc)
	nextPage := 0
	totalPages := int(math.Ceil(float64(products.Count) / float64(perPage)))
	if page < totalPages {
		nextPage = page + 1
	}

	RenderClientWithLayout(
		templates.ClientProductsChunk(cards, nextPage, query),
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

	RenderClientWithLayout(
		templates.ClientCheckoutPage(c.getCartCount(sess) == 0),
		w,
		r,
		c.getCartCount(sess),
		"carrinho",
	)
}

func makeCardData(
	products []store.Product,
	table store.PriceTable,
	ptSvc services.PriceTableService,
) []store.CardData {
	cards := make([]store.CardData, 0, len(products))
	for _, p := range products {
		variants := make([]store.VariantData, 0, len(p.Variants))
		for _, v := range p.Variants {
			attrs := make([]store.AttrData, 0, len(v.Attributes))
			for _, a := range v.Attributes {
				attrs = append(attrs, store.AttrData{
					Name:  a.AttributeValue.Attribute.Name,
					Value: a.AttributeValue.Value,
				})
			}
			variants = append(variants, store.VariantData{
				ID:        v.ID,
				Price:     ptSvc.Apply(v.CostPrice, &table),
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
	return cards
}
