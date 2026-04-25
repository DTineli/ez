package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/forms"
	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
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
| Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) GetProductForm(w http.ResponseWriter, r *http.Request) {
	Render(templates.ProductForm(forms.New(nil), false, nil), r, w)
}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	form, err := validateProductForm(r)

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		ShowToast(w, "Erros de validação", "error")
		_ = Render(templates.ProductForm(form, false, nil), r, w)
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
	}

	if err := p.productStore.CreateProduct(product); err != nil {
		if isDuplicateError(err) {
			ShowToast(w, "Erros de validação", "error")
			form.Errors.Add("sku", "Este SKU já está em uso.")
			_ = Render(templates.ProductForm(form, false, nil), r, w)
			return
		}
		ShowToast(w, "Erro ao cadastar produto", "error")
		_ = Render(templates.ProductForm(form, false, nil), r, w)
		return
	}

	ShowToast(w, "Produto Cadastrado", "success")
	form.Set("ID", strconv.Itoa(int(product.ID)))
	_ = Render(templates.ProductForm(form, true, nil), r, w)
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

	variants, _ := p.productStore.FindVariantsByProduct(uint(id), sess.TenantID)

	if !form.Valid() {
		ShowToast(w, "Erro ao salvar produto", "error")

		_ = Render(templates.ProductForm(form, true, variants), r, w)
		return
	}

	fields := map[string]any{
		"name":             form.Get("name"),
		"full_description": form.Get("description"),
		"status":           true,
		"uom":              store.UOM(form.Get("uom")),
		"ean":              form.Get("ean"),
		"ncm":              form.Get("ncm"),
	}

	err = p.productStore.UpdateFields(uint(id), sess.TenantID, fields)
	if err != nil {
		ShowToast(w, "Erro ao salvar produto", "error")
		_ = Render(templates.ProductForm(form, true, variants), r, w)
	}

	ShowToast(w, "Alteracoes Salvas", "success")
	_ = Render(templates.ProductForm(form, true, variants), r, w)
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

	_ = Render(templates.ProductForm(form, true, product.Variants), r, w)
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

/*
|--------------------------------------------------------------------------
| Variant Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) GetVariantForm(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	Render(templates.NewVariantForm(productID), r, w)
}

func (p *ProductHandler) CancelVariantForm(w http.ResponseWriter, r *http.Request) {
	productID := chi.URLParam(r, "id")
	Render(templates.NewVariantTrigger(productID), r, w)
}

func (p *ProductHandler) GetEditVariantRow(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	variantID, err := strconv.ParseUint(chi.URLParam(r, "variantID"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	variant, err := p.productStore.GetVariant(uint(variantID), sess.TenantID)
	if err != nil {
		http.Error(w, "variação não encontrada", http.StatusNotFound)
		return
	}

	Render(templates.VariantEditRow(*variant, chi.URLParam(r, "id")), r, w)
}

func (p *ProductHandler) GetVariantRow(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	variantID, err := strconv.ParseUint(chi.URLParam(r, "variantID"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	variant, err := p.productStore.GetVariant(uint(variantID), sess.TenantID)

	if err != nil {
		http.Error(w, "variação não encontrada", http.StatusNotFound)
		return
	}

	Render(templates.VariantRow(*variant, chi.URLParam(r, "id")), r, w)
}

func (p *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	variantID, err := strconv.ParseUint(chi.URLParam(r, "variantID"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	costPrice, _ := strconv.ParseFloat(r.FormValue("cost_price"), 64)
	currentStock, _ := strconv.Atoi(r.FormValue("current_stock"))
	minimumStock, _ := strconv.Atoi(r.FormValue("minimum_stock"))

	fields := map[string]any{
		"cost_price":    costPrice,
		"current_stock": currentStock,
		"minimum_stock": minimumStock,
	}

	if err := p.productStore.UpdateVariantFields(uint(variantID), sess.TenantID, fields); err != nil {
		ShowToast(w, "Erro ao salvar variação", "error")
		variant, _ := p.productStore.GetVariant(uint(variantID), sess.TenantID)
		Render(templates.VariantEditRow(*variant, chi.URLParam(r, "id")), r, w)
		return
	}

	ShowToast(w, "Variação salva", "success")
	variant, _ := p.productStore.GetVariant(uint(variantID), sess.TenantID)
	Render(templates.VariantRow(*variant, chi.URLParam(r, "id")), r, w)
}

func (p *ProductHandler) PostVariant(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	productID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	costPrice, _ := strconv.ParseFloat(r.FormValue("cost_price"), 64)
	currentStock, _ := strconv.Atoi(r.FormValue("current_stock"))
	minimumStock, _ := strconv.Atoi(r.FormValue("minimum_stock"))

	variant := &store.Variant{
		SKU:          r.FormValue("sku"),
		CostPrice:    costPrice,
		CurrentStock: currentStock,
		MinimumStock: minimumStock,
		ProductID:    uint(productID),
		TenantID:     sess.TenantID,
	}

	if err := p.productStore.CreateVariant(variant); err != nil {
		ShowToast(w, "Erro ao cadastrar variação", "error")
		variants, _ := p.productStore.FindVariantsByProduct(uint(productID), sess.TenantID)
		Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
		return
	}

	ShowToast(w, "Variação cadastrada", "success")
	variants, _ := p.productStore.FindVariantsByProduct(uint(productID), sess.TenantID)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}
