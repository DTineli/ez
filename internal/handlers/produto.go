package handlers

import (
	"net/http"

	"github.com/DTineli/ez/internal/forms"
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
		err := templates.ProductForm(forms.New(r.PostForm)).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	err := templates.Layout(
		templates.ProductForm(forms.New(r.PostForm)),
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

type CreateProductDTO struct {
	Name         string  `schema:"name" validate:"required,min=3"`
	SKU          string  `schema:"sku" validate:"required, min=4"`
	EAN          string  `schema:"ean" validate:"min=13, max=13"`
	Description  string  `schema:"description" validate:"min 5"`
	UOM          string  `schema:"uom" validate:"required"`
	NCM          string  `schema:"ncm"`
	CostPrice    float64 `schema:"cost_price" validate:"required, gte=0"`
	Stock        int     `schema:"current_stock" validate:"gte=0"`
	MinimumStock int     `schema:"minimum_stock" validate:"get=0"`
	Weight       float64 `schema:"weight" validate:"gte=0"`
	Height       float64 `schema:"height" validate:"gte=0"`
	Width        float64 `schema:"width" validate:"gte=0"`
	Length       float64 `schema:"length" validate:"gte=0"`
}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("name", "sku")

	form.MaxLength("name", 255)
	form.MinLength("name", 4)

	form.MaxLength("uom", 2)
	form.MinLength("uom", 2)

	form.MaxLength("Description", 255)

	form.MaxLength("sku", 25)
	form.MinLength("sku", 4)

	form.IsFloat("cost_price")
	form.IsFloat("weight")
	form.IsFloat("height")
	form.IsFloat("width")
	form.IsFloat("Length")
	form.IsFloat("weight")

	form.IsInt("current_stock")
	form.IsInt("minimum_stock")
	form.IsInt("ean")

	if !form.Valid() {
		err := templates.ProductForm(form).Render(r.Context(), w)

		if err != nil {
			http.Error(w, "Error Creating Product", http.StatusInternalServerError)
		}

		return
	}

	_ = m.GetSessionFromContext(r)

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
