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

/*
|--------------------------------------------------------------------------
| Handlers
|--------------------------------------------------------------------------
*/

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

	ShowToast(w, "Produto Cadastrado com Sucesso", "success")
	w.Header().
		Set("HX-Redirect", "/admin/produtos/"+strconv.Itoa(int(product.ID)))
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

	fromPage := r.FormValue("ctx") == "page"

	if err := p.productStore.CreateAttributeValue(av); err != nil {
		if isDuplicateError(err) {
			ShowToast(w, "Valor já cadastrado neste atributo", "error")
		} else {
			ShowToast(w, "Erro ao cadastrar valor", "error")
		}
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		if fromPage {
			Render(templates.AttributesPage(attrs), r, w)
		} else {
			Render(templates.AttributesSection(attrs), r, w)
		}
		return
	}

	ShowToast(w, "Valor cadastrado", "success")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	if fromPage {
		Render(templates.AttributesPage(attrs), r, w)
	} else {
		Render(templates.AttributesSection(attrs), r, w)
	}
}

func (p *ProductHandler) GetAttributeForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributeManagementSectionWithForm(attrs), r, w)
}

func (p *ProductHandler) CancelAttributeForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributeManagementSection(attrs), r, w)
}

func (p *ProductHandler) PostNewAttribute(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	name := strings.ToLower(strings.TrimSpace(r.FormValue("name")))
	if name == "" {
		http.Error(w, "nome é obrigatório", http.StatusUnprocessableEntity)
		return
	}

	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	for _, existing := range attrs {
		if strings.EqualFold(
			strings.TrimSpace(existing.Name),
			strings.TrimSpace(name),
		) {
			ShowToast(w, "Este atributo já existe", "error")
			Render(templates.AttributeManagementSectionWithForm(attrs), r, w)
			return
		}
	}

	attr := &store.Attribute{
		Name:     name,
		TenantID: sess.TenantID,
	}

	fromPage := r.FormValue("ctx") == "page"

	if err := p.productStore.CreateAttribute(attr); err != nil {
		ShowToast(w, "Erro ao cadastrar atributo", "error")
		attrs2, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		if fromPage {
			Render(templates.AttributesPage(attrs2), r, w)
		} else {
			Render(templates.AttributeManagementSectionWithForm(attrs2), r, w)
		}
		return
	}

	ShowToast(w, "Atributo cadastrado", "success")
	attrs = append(attrs, *attr)
	if fromPage {
		Render(templates.AttributesPage(attrs), r, w)
	} else {
		Render(templates.AttributeManagementSection(attrs), r, w)
	}
}

func (p *ProductHandler) GetAttributesPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributesPage(attrs), r, w)
}

func (p *ProductHandler) DeleteAttribute(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	inUse, err := p.productStore.AttributeInUse(uint(id), sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao verificar atributo", "error")
		return
	}
	if inUse {
		ShowToast(
			w,
			"Atributo em uso por variações, não pode ser deletado",
			"error",
		)
		return
	}

	if err := p.productStore.DeleteAttribute(uint(id), sess.TenantID); err != nil {
		ShowToast(w, "Erro ao deletar atributo", "error")
		return
	}

	ShowToast(w, "Atributo deletado", "success")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
	Render(templates.AttributesPage(attrs), r, w)
}

/*
|--------------------------------------------------------------------------
| Variant Handlers
|--------------------------------------------------------------------------
*/

func (p *ProductHandler) GetVariantForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	productID := chi.URLParam(r, "id")
	attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)

	productSKU := r.URL.Query().Get("sku")
	if productSKU == "" {
		http.Error(w, "sku inválido", http.StatusBadRequest)
		return
	}

	Render(templates.NewVariantForm(productID, productSKU, attrs), r, w)
}

func (p *ProductHandler) CancelVariantForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}

func (p *ProductHandler) GetEditVariantRow(
	w http.ResponseWriter,
	r *http.Request,
) {
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

func (p *ProductHandler) BulkUpdateVariants(
	w http.ResponseWriter,
	r *http.Request,
) {
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

	addStockStr := r.FormValue("add_stock")
	setMinStockStr := r.FormValue("add_min_stock")
	setPriceStr := r.FormValue("set_price")

	if addStockStr == "" && setPriceStr == "" && setMinStockStr == "" {
		variants, _ := p.productStore.FindVariantsByProduct(
			uint(productID),
			sess.TenantID,
		)
		Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
		return
	}

	variants, err := p.productStore.FindVariantsByProduct(
		uint(productID),
		sess.TenantID,
	)
	if err != nil {
		http.Error(
			w,
			"erro ao buscar variações",
			http.StatusInternalServerError,
		)
		return
	}

	setStock, _ := strconv.Atoi(addStockStr)
	setMinStock, _ := strconv.Atoi(setMinStockStr)
	setPrice, _ := strconv.ParseFloat(setPriceStr, 64)

	for _, v := range variants {
		fields := map[string]any{}
		if addStockStr != "" {
			fields["current_stock"] = setStock
		}
		if setPriceStr != "" {
			fields["cost_price"] = setPrice
		}

		if setMinStockStr != "" {
			fields["minimum_stock"] = setMinStock
		}
		if len(fields) > 0 {
			_ = p.productStore.UpdateVariantFields(v.ID, sess.TenantID, fields)
		}
	}

	ShowToast(w, "Variações atualizadas", "success")
	variants, _ = p.productStore.FindVariantsByProduct(
		uint(productID),
		sess.TenantID,
	)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}

func hasSales(variantID uint) bool {
	return false
}

func (p *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	variantID, err := strconv.ParseUint(chi.URLParam(r, "variantID"), 10, 64)

	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	if !hasSales(uint(variantID)) {
		p.productStore.DeleteVariant(uint(variantID), sess.TenantID)
	}

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

	if defaultVariant, err := p.productStore.FindDefaultVariant(uint(productID), sess.TenantID); err == nil &&
		defaultVariant != nil {
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
		variants, _ := p.productStore.FindVariantsByProduct(
			uint(productID),
			sess.TenantID,
		)
		Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
		return
	}

	var attributeValueIDs []uint
	for key, vals := range r.Form {
		if !strings.HasPrefix(key, "attr_name_") || len(vals) == 0 ||
			strings.TrimSpace(vals[0]) == "" {
			continue
		}
		idxStr := strings.TrimPrefix(key, "attr_name_")
		name := strings.TrimSpace(vals[0])
		value := strings.TrimSpace(r.FormValue("attr_value_" + idxStr))
		if value == "" {
			continue
		}
		attr, err := p.productStore.FindOrCreateAttribute(name, sess.TenantID)
		if err != nil {
			continue
		}
		av, err := p.productStore.FindOrCreateAttributeValue(value, attr.ID)
		if err != nil {
			continue
		}
		attributeValueIDs = append(attributeValueIDs, av.ID)
	}

	if len(attributeValueIDs) == 0 {
		_ = p.productStore.DeleteVariant(variant.ID, sess.TenantID)
		ShowToast(w, "Informe ao menos um valor de atributo", "error")
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		parentSku := ""
		if prod, err := p.productStore.GetProduct(uint(productID)); err == nil &&
			prod != nil {
			parentSku = prod.SKU
		}
		Render(
			templates.NewVariantForm(chi.URLParam(r, "id"), parentSku, attrs),
			r,
			w,
		)
		return
	}

	if len(attributeValueIDs) > 0 {
		if err := p.productStore.SetVariantAttributes(variant.ID, attributeValueIDs); err != nil {
			ShowToast(
				w,
				"Variação cadastrada, mas erro ao salvar atributos",
				"error",
			)
			variants, _ := p.productStore.FindVariantsByProduct(
				uint(productID),
				sess.TenantID,
			)
			Render(
				templates.VariantsSection(variants, chi.URLParam(r, "id")),
				r,
				w,
			)
			return
		}
	}

	ShowToast(w, "Variação cadastrada", "success")
	variants, _ := p.productStore.FindVariantsByProduct(
		uint(productID),
		sess.TenantID,
	)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}
