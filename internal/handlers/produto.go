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
	Render(templates.ProductForm(forms.New(nil), false, nil, nil), r, w)
}

func (p *ProductHandler) PostNewProduct(w http.ResponseWriter, r *http.Request) {
	form, err := validateProductForm(r)

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		_ = Render(templates.ProductForm(form, false, nil, nil), r, w)
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
			form.Errors.Add("sku", "Este SKU já está em uso.")
		} else {
			form.Errors.Add("general", "Erro ao cadastrar produto. Tente novamente.")
		}
		_ = Render(templates.ProductForm(form, false, nil, nil), r, w)
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

	ShowToast(w, "Produto Cadastrado com Sucesso", "success")
	w.Header().Set("HX-Redirect", "/admin/produtos/"+strconv.Itoa(int(product.ID)))
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
		_ = Render(templates.ProductForm(form, true, variants, attrs), r, w)
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
		_ = Render(templates.ProductForm(form, true, variants, attrs), r, w)
		return
	}

	ShowToast(w, "Alteracoes Salvas", "success")
	_ = Render(templates.ProductForm(form, true, variants, attrs), r, w)
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

	_ = Render(templates.ProductForm(form, true, product.Variants, attrs), r, w)
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
| Attributes Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) PostAddValue(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	attributeID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	value := r.FormValue("value")
	if value == "" {
		http.Error(w, "valor é obrigatório", http.StatusUnprocessableEntity)
		return
	}

	if _, err := p.productStore.GetAttribute(uint(attributeID), sess.TenantID); err != nil {
		http.Error(w, "atributo não encontrado", http.StatusNotFound)
		return
	}

	av := &store.AttributeValue{
		Value:       value,
		AttributeID: uint(attributeID),
	}

	if err := p.productStore.CreateAttributeValue(av); err != nil {
		if isDuplicateError(err) {
			ShowToast(w, "Valor já cadastrado neste atributo", "error")
		} else {
			ShowToast(w, "Erro ao cadastrar valor", "error")
		}
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		Render(templates.AttributesSection(attrs), r, w)
		return
	}

	ShowToast(w, "Valor cadastrado", "success")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributesSection(attrs), r, w)
}

func (p *ProductHandler) PostNewAttribute(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "nome é obrigatório", http.StatusUnprocessableEntity)
		return
	}

	attr := &store.Attribute{
		Name:     name,
		TenantID: sess.TenantID,
	}

	if err := p.productStore.CreateAttribute(attr); err != nil {
		ShowToast(w, "Erro ao cadastrar atributo", "error")
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		Render(templates.AttributesSection(attrs), r, w)
		return
	}

	ShowToast(w, "Atributo cadastrado", "success")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributesSection(attrs), r, w)
}

/*
|--------------------------------------------------------------------------
| Variant Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) GetVariantForm(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	productID := chi.URLParam(r, "id")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.NewVariantForm(productID, attrs), r, w)
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

	if defaultVariant, err := p.productStore.FindDefaultVariant(uint(productID), sess.TenantID); err == nil && defaultVariant != nil {
		_ = p.productStore.DeleteVariant(defaultVariant.ID, sess.TenantID)
	}

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

	var attributeValueIDs []uint
	for _, raw := range r.Form["attribute_value_ids"] {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
			attributeValueIDs = append(attributeValueIDs, uint(id))
		}
	}

	if len(attributeValueIDs) == 0 {
		_ = p.productStore.DeleteVariant(variant.ID, sess.TenantID)
		ShowToast(w, "Selecione ao menos um atributo", "error")
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		Render(templates.NewVariantForm(chi.URLParam(r, "id"), attrs), r, w)
		return
	}

	if len(attributeValueIDs) > 0 {
		if err := p.productStore.SetVariantAttributes(variant.ID, attributeValueIDs); err != nil {
			ShowToast(w, "Variação cadastrada, mas erro ao salvar atributos", "error")
			variants, _ := p.productStore.FindVariantsByProduct(uint(productID), sess.TenantID)
			Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
			return
		}
	}

	ShowToast(w, "Variação cadastrada", "success")
	variants, _ := p.productStore.FindVariantsByProduct(uint(productID), sess.TenantID)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}
