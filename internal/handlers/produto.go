package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/DTineli/ez/internal/forms"
	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	productStore    store.ProductStore
	priceTableStore store.PriceTableStore
}

func NewProductHandler(productDB store.ProductStore, priceTableDB store.PriceTableStore) *ProductHandler {
	return &ProductHandler{
		productStore:    productDB,
		priceTableStore: priceTableDB,
	}
}

/*
|--------------------------------------------------------------------------
| Helpers
|--------------------------------------------------------------------------
*/

func Render(c templ.Component, r *http.Request, w http.ResponseWriter) error {
	if r.Header.Get("HX-Request") == "true" {
		return c.Render(r.Context(), w)
	}

	return templates.
		Layout(c, "Ez", true, "").
		Render(r.Context(), w)
}

func ShowToast(w http.ResponseWriter, message string, toastType string) {
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{
	"showToast": {
		"type": "%v",
		"message": "%v"
	}
}`, toastType, message))
	w.WriteHeader(http.StatusOK)
}

func validateProductForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)

	form.Required("name", "sku")

	form.MaxLength("name", 255)
	form.MinLength("name", 4)

	form.MaxLength("uom", 2)
	form.MinLength("uom", 2)

	form.MaxLength("description", 15000)

	form.MaxLength("sku", 25)
	form.MinLength("sku", 4)

	form.IsFloat("cost_price")
	form.IsFloat("weight")
	form.IsFloat("height")
	form.IsFloat("width")
	form.IsFloat("length")

	form.IsInt("current_stock")
	form.IsInt("minimum_stock")
	form.IsInt("ean")
	form.Set("ncm", strings.ReplaceAll(form.Get("ncm"), ".", ""))
	form.IsInt("ncm")

	return form, nil
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "Duplicate")
}

func mapProductToForm(p *store.Product) *forms.Form {
	form := forms.New(nil)

	form.Set("ID", strconv.Itoa(int(p.ID)))
	form.Set("name", p.Name)
	form.Set("sku", p.SKU)
	form.Set("uom", string(p.UOM))
	form.Set("description", p.FullDescription)

	form.Set("cost_price", strconv.FormatFloat(p.CostPrice, 'f', 2, 64))
	form.Set("weight", strconv.FormatFloat(p.Weight, 'f', 2, 64))
	form.Set("height", strconv.FormatFloat(p.HeightCm, 'f', 2, 64))
	form.Set("length", strconv.FormatFloat(p.LengthCm, 'f', 2, 64))
	form.Set("width", strconv.FormatFloat(p.WidthCm, 'f', 2, 64))

	form.Set("ean", p.EAN)
	ncm := p.NCM
	if len(ncm) == 8 {
		ncm = ncm[:4] + "." + ncm[4:6] + "." + ncm[6:]
	}
	form.Set("ncm", ncm)

	form.Set("minimum_stock", strconv.Itoa(p.MinimumStock))
	form.Set("current_stock", strconv.Itoa(p.CurrentStock))

	return form
}

/*
|--------------------------------------------------------------------------
| Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) GetProductForm(w http.ResponseWriter, r *http.Request) {
	Render(templates.ProductForm(forms.New(nil), false), r, w)
}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	form, err := validateProductForm(r)
	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		ShowToast(w, "Erros de validação", "error")
		_ = Render(templates.ProductForm(form, false), r, w)
		return
	}

	sess := m.GetSessionFromContext(r)

	product := &store.Product{
		TenantID:        sess.TenantID,
		SKU:             form.Get("sku"),
		Name:            form.Get("name"),
		FullDescription: form.Get("description"),
		Status:          true,
		UOM:             store.UOM(form.Get("uom")),
		EAN:             form.Get("ean"),
		NCM:             form.Get("ncm"),
		CostPrice:       form.IsFloat("cost_price"),
		WidthCm:         form.IsFloat("width"),
		Weight:          form.IsFloat("weight"),
		HeightCm:        form.IsFloat("height"),
		LengthCm:        form.IsFloat("length"),
		MinimumStock:    form.IsInt("minimum_stock"),
		CurrentStock:    form.IsInt("current_stock"),
	}

	if err := p.productStore.CreateProduct(product); err != nil {
		if isDuplicateError(err) {
			ShowToast(w, "Erros de validação", "error")
			form.Errors.Add("sku", "Este SKU já está em uso.")
			_ = Render(templates.ProductForm(form, false), r, w)
			return
		}
		ShowToast(w, "Erro ao cadastar produto", "error")
		_ = Render(templates.ProductForm(form, false), r, w)
		return
	}

	ShowToast(w, "Produto Cadastrado", "success")
	form.Set("ID", strconv.Itoa(int(product.ID)))
	_ = Render(templates.ProductForm(form, true), r, w)
}

func (p *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	form, err := validateProductForm(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	form.Set("ID", strconv.Itoa(int(id)))

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		ShowToast(w, "Erro ao salvar produto", "error")

		_ = Render(templates.ProductForm(form, true), r, w)
		return
	}

	fields := map[string]any{
		"name":             form.Get("name"),
		"full_description": form.Get("description"),
		"status":           true,
		"uom":              store.UOM(form.Get("uom")),
		"ean":              form.Get("ean"),
		"ncm":              form.Get("ncm"),
		"cost_price":       form.IsFloat("cost_price"),
		"width_cm":         form.IsFloat("width"),
		"weight":           form.IsFloat("weight"),
		"height_cm":        form.IsFloat("height"),
		"length_cm":        form.IsFloat("length"),
		"minimum_stock":    form.IsInt("minimum_stock"),
		"current_stock":    form.IsInt("current_stock"),
	}

	err = p.productStore.UpdateFields(uint(id), sess.TenantID, fields)
	if err != nil {
		ShowToast(w, "Erro ao salvar produto", "error")
		_ = Render(templates.ProductForm(form, true), r, w)
	}

	ShowToast(w, "Alteracoes Salvas", "success")
	_ = Render(templates.ProductForm(form, true), r, w)
}

func (p *ProductHandler) GetEditPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	product, err := p.productStore.GetProduct(uint(id))
	if err != nil || product.TenantID != sess.TenantID {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	form := mapProductToForm(product)

	_ = Render(templates.ProductForm(form, true), r, w)
}

func (p *ProductHandler) GetProductPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	page := 1
	perPage := 10
	if strPage := r.URL.Query().Get("page"); strPage != "" {
		if p, err := strconv.Atoi(strPage); err == nil && p > 0 {
			page = p
		}
	}

	results, err := p.productStore.FindAllByUserWithFilters(sess.TenantID, store.ProductFilters{
		Page:    page,
		PerPage: perPage,
		SKU:     r.URL.Query().Get("sku"),
		Name:    r.URL.Query().Get("name"),
	})

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error listing Product", http.StatusInternalServerError)
		return
	}

	err = Render(templates.ProductsPage(store.GetProductPageParams{
		Products:   results.Results,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int(math.Floor(float64(results.Count) / float64(perPage))),
	}), r, w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: implementar
}
