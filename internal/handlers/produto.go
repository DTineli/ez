package handlers

import (
	"fmt"
	"net/http"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/schema"
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
	Length       float64 `schema:"Length" validate:"gte=0"`
}

func validateFormValues(r *http.Request) map[string]string {
	erros := make(map[string]string)
	if r.FormValue("name") == "" {
		erros["name"] = "Nome Obrigatorio"
	}

	return erros
}

var decoder = schema.NewDecoder()

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	var form CreateProductDTO

	err := r.ParseForm()

	if err := decoder.Decode(&form, r.PostForm); err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 422)
		return
	}

	if err := validate.Struct(form); err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	_ = m.GetSessionFromContext(r)

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
