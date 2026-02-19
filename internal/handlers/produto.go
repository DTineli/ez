package handlers

import (
	"net/http"
	"strconv"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type ProductHandler struct {
	productStore store.ProductStore
}

func NewProductHandler(db store.ProductStore) *ProductHandler {
	return &ProductHandler{
		productStore: db,
	}
}

func (p *ProductHandler) GetProductForm(w http.ResponseWriter, r *http.Request) {
	var is_hxRequest = r.Header.Get("HX-Request") == "true"

	if is_hxRequest {
		err := templates.ProductForm().Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(
		templates.ProductForm(),
		"Ez",
		true,
		"",
	).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {

}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	stock, err := strconv.Atoi(r.FormValue("stock"))

	name := r.FormValue("name")
	sku := r.FormValue("sku")

	if name == "" {
		http.Error(w, "Nome é obrigatorio", http.StatusBadRequest)
		return
	}

	if sku == "" {
		http.Error(w, "sku é obrigatorio", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error convert string", http.StatusBadRequest)
		return
	}

	sess := m.GetSessionFromContext(r)

	product := &store.Product{
		TenantID:     sess.TenantID,
		Name:         name,
		SKU:          sku,
		CostPrice:    price,
		CurrentStock: stock,
	}

	err = p.productStore.CreateProduct(product)
	if err != nil {
		http.Error(w, "Error Creating Product", http.StatusInternalServerError)
	}

	w.Header().Set(HXRedirect, "/produtos")
	w.WriteHeader(http.StatusOK)
}

func (p *ProductHandler) GetProductPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	var is_hxRequest = r.Header.Get("HX-Request") == "true"

	produtos, err := p.productStore.FindAllByUser(sess.TenantID)

	if err != nil {
		http.Error(w, "Error listing Product", http.StatusInternalServerError)
	}

	if is_hxRequest {
		err := templates.ProductsPage(produtos).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err = templates.Layout(
		templates.ProductsPage(produtos),
		"Ez",
		true,
		"",
	).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
