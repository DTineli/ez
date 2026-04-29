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

func NewProductHandler(
	productDB store.ProductStore,
	priceTableDB store.PriceTableStore,
) *ProductHandler {
	return &ProductHandler{
		productStore:    productDB,
		priceTableStore: priceTableDB,
	}
}

func (p *ProductHandler) GetProductForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(
		templates.ProductForm(forms.New(nil), false, false, nil, attrs),
		r,
		w,
	)
}

func (p *ProductHandler) PostNewProduct(
	w http.ResponseWriter,
	r *http.Request,
) {
	form, err := validateProductForm(r)

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	sess := m.GetSessionFromContext(r)

	if !form.Valid() {
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		_ = Render(templates.ProductForm(form, false, false, nil, attrs), r, w)
		return
	}

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
			form.Errors.Add("sku", "Este SKU já está em uso.")
		} else {
			form.Errors.Add("general", "Erro ao cadastrar produto. Tente novamente.")
		}
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		_ = Render(templates.ProductForm(form, false, false, nil, attrs), r, w)
		return
	}

	costPrice, _ := strconv.ParseFloat(form.Get("cost_price"), 64)
	currentStock, _ := strconv.Atoi(form.Get("current_stock"))
	minimumStock, _ := strconv.Atoi(form.Get("minimum_stock"))

	defaultVariant := &store.Variant{
		SKU:          product.SKU,
		ProductID:    product.ID,
		CostPrice:    costPrice,
		CurrentStock: currentStock,
		MinimumStock: minimumStock,
		TenantID:     sess.TenantID,
		IsDefault:    true,
	}
	_ = p.productStore.CreateVariant(defaultVariant)

	HXLocation(
		w,
		"/admin/produtos/"+strconv.Itoa(int(product.ID)),
		"Produto Cadastrado com Sucesso",
		"success",
	)
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
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)

	if !form.Valid() {
		_ = Render(
			templates.ProductForm(form, true, false, variants, attrs),
			r,
			w,
		)
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
		form.Errors.Add("general", "Erro ao salvar produto. Tente novamente.")
		_ = Render(
			templates.ProductForm(form, true, false, variants, attrs),
			r,
			w,
		)
		return
	}

	ShowToast(w, "Alteracoes Salvas", "success")
	_ = Render(templates.ProductForm(form, true, false, variants, attrs), r, w)
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
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)

	_ = Render(
		templates.ProductForm(form, true, false, product.Variants, attrs),
		r,
		w,
	)
}

func (p *ProductHandler) GetProductPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	page := 1
	perPage := 10
	if strPage := r.URL.Query().Get("page"); strPage != "" {
		if p, err := strconv.Atoi(strPage); err == nil && p > 0 {
			page = p
		}
	}

	results, err := p.productStore.FindAllByUserWithFilters(
		sess.TenantID,
		store.ProductFilters{
			Page:    page,
			PerPage: perPage,
			SKU:     r.URL.Query().Get("sku"),
			Name:    r.URL.Query().Get("name"),
		},
	)

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
		http.Error(
			w,
			"Error rendering template",
			http.StatusInternalServerError,
		)
	}
}

func (p *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// TODO: implementar
}
