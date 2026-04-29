package handlers

import (
	"net/http"
	"strconv"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/go-chi/chi/v5"
)

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

	defaultCostPrice, _ := strconv.ParseFloat(r.URL.Query().Get("default_cost_price"), 64)

	Render(templates.NewVariantForm(productID, productSKU, attrs, defaultCostPrice), r, w)
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
		http.Error(w, "erro ao buscar variações", http.StatusInternalServerError)
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

func (p *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	variantID, err := strconv.ParseUint(chi.URLParam(r, "variantID"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	p.productStore.DeleteVariant(uint(variantID), sess.TenantID)
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

	renderVariantForm := func(msg string) {
		ShowToast(w, msg, "error")
		attrs, _ := p.productStore.FindAttributesByTenant(sess.TenantID)
		parentSku := ""
		if prod, err := p.productStore.GetProduct(uint(productID)); err == nil &&
			prod != nil {
			parentSku = prod.SKU
		}
		variants, _ := p.productStore.FindVariantsByProduct(uint(productID), sess.TenantID)
		var defaultCostPrice float64
		if len(variants) > 0 {
			defaultCostPrice = variants[0].CostPrice
		}
		Render(
			templates.VariantsSectionWithNewForm(variants, chi.URLParam(r, "id"), parentSku, attrs, defaultCostPrice),
			r,
			w,
		)
	}

	sku := strings.TrimSpace(r.FormValue("sku"))
	if sku == "" {
		renderVariantForm("SKU é obrigatório")
		return
	}

	attrInputs, err := parseVariantAttributes(r)
	if err != nil {
		renderVariantForm(err.Error())
		return
	}
	if len(attrInputs) == 0 {
		renderVariantForm("Informe ao menos um valor de atributo")
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
		SKU:          sku,
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
	for _, ai := range attrInputs {
		attr, err := p.productStore.FindOrCreateAttribute(ai.Name, sess.TenantID)
		if err != nil {
			continue
		}
		av, err := p.productStore.FindOrCreateAttributeValue(ai.Value, attr.ID)
		if err != nil {
			continue
		}
		attributeValueIDs = append(attributeValueIDs, av.ID)
	}

	if err := p.productStore.SetVariantAttributes(variant.ID, attributeValueIDs); err != nil {
		_ = p.productStore.DeleteVariant(variant.ID, sess.TenantID)
		renderVariantForm("Erro ao salvar atributos, variação não cadastrada")
		return
	}

	ShowToast(w, "Variação cadastrada", "success")
	variants, _ := p.productStore.FindVariantsByProduct(
		uint(productID),
		sess.TenantID,
	)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}
