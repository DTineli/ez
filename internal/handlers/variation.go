package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/DTineli/ez/internal/validate"
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

	defaultCostPrice, _ := strconv.ParseFloat(
		r.URL.Query().Get("default_cost_price"),
		64,
	)

	Render(
		templates.NewVariantForm(
			productID,
			productSKU,
			attrs,
			defaultCostPrice,
		),
		r,
		w,
	)
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

	existing, err := p.productStore.GetVariant(uint(variantID), sess.TenantID)
	if err != nil {
		http.Error(w, "variação não encontrada", http.StatusNotFound)
		return
	}

	costPrice, _ := strconv.ParseFloat(r.FormValue("cost_price"), 64)
	currentStock, _ := strconv.Atoi(r.FormValue("current_stock"))
	ean, err := validate.EAN(r.FormValue("ean"))
	status := r.FormValue("status") == "true"

	if err != nil {
		ShowToast(w, err.Error(), "error")
		Render(templates.VariantEditRow(*existing, chi.URLParam(r, "id")), r, w)
		return
	}

	fields := map[string]any{
		"cost_price":    costPrice,
		"current_stock": currentStock,
		"ean":           ean,
		"status":        status,
	}

	if err := p.productStore.UpdateVariantFields(uint(variantID), sess.TenantID, fields); err != nil {
		ShowToast(w, "Erro ao salvar variação", "error")
		Render(templates.VariantEditRow(*existing, chi.URLParam(r, "id")), r, w)
		return
	}

	p.productStore.RecalcularStatusProduto(existing.ProductID, sess.TenantID)

	existing.CostPrice = costPrice
	existing.CurrentStock = currentStock
	existing.EAN = ean
	existing.Status = status
	ShowToast(w, "Variação salva", "success")
	Render(templates.VariantRow(*existing, chi.URLParam(r, "id")), r, w)
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
	setPriceStr := r.FormValue("set_price")

	if addStockStr == "" && setPriceStr == "" {
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
	setPrice, _ := strconv.ParseFloat(setPriceStr, 64)

	for _, v := range variants {
		fields := map[string]any{}
		if addStockStr != "" {
			fields["current_stock"] = setStock
		}
		if setPriceStr != "" {
			fields["cost_price"] = setPrice
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
		variants, _ := p.productStore.FindVariantsByProduct(
			uint(productID),
			sess.TenantID,
		)
		var defaultCostPrice float64
		if len(variants) > 0 {
			defaultCostPrice = variants[0].CostPrice
		}
		Render(
			templates.VariantsSectionWithNewForm(
				variants,
				chi.URLParam(r, "id"),
				parentSku,
				attrs,
				defaultCostPrice,
			),
			r,
			w,
		)
	}

	skuBase := strings.TrimSpace(r.FormValue("sku"))
	if skuBase == "" {
		renderVariantForm("SKU é obrigatório")
		return
	}

	axes, err := parseVariantAxes(r)
	if err != nil {
		renderVariantForm(err.Error())
		return
	}
	if len(axes) == 0 {
		renderVariantForm("Informe ao menos um atributo e valor")
		return
	}

	combos := cartesianCombos(axes)
	if len(combos) > maxVariantCombos {
		renderVariantForm(fmt.Sprintf("Muitas combinações (%d), limite é %d. Reduza os valores.", len(combos), maxVariantCombos))
		return
	}

	costPrice, _ := strconv.ParseFloat(r.FormValue("cost_price"), 64)
	currentStock, _ := strconv.Atoi(r.FormValue("current_stock"))

	// Resolve Attribute/AttributeValue uma vez por valor único de cada eixo.
	valueIDs := make([]map[string]uint, len(axes))
	for i, axis := range axes {
		attr, err := p.productStore.FindOrCreateAttribute(axis.Name, sess.TenantID)
		if err != nil {
			renderVariantForm("Erro ao resolver atributo: " + axis.Name)
			return
		}
		valueIDs[i] = make(map[string]uint, len(axis.Values))
		for _, value := range axis.Values {
			av, err := p.productStore.FindOrCreateAttributeValue(value, attr.ID)
			if err != nil {
				renderVariantForm("Erro ao resolver valor: " + value)
				return
			}
			valueIDs[i][value] = av.ID
		}
	}

	inputs := make([]store.VariantGenInput, 0, len(combos))
	for _, combo := range combos {
		attributeValueIDs := make([]uint, len(combo))
		for i, value := range combo {
			attributeValueIDs[i] = valueIDs[i][value]
		}
		inputs = append(inputs, store.VariantGenInput{
			SKU:               skuBase + "-" + strings.Join(combo, "-"),
			CostPrice:         costPrice,
			CurrentStock:      currentStock,
			AttributeValueIDs: attributeValueIDs,
		})
	}

	if defaultVariant, err := p.productStore.FindDefaultVariant(uint(productID), sess.TenantID); err == nil &&
		defaultVariant != nil {
		_ = p.productStore.DeleteVariant(defaultVariant.ID, sess.TenantID)
	}

	if _, err := p.productStore.CreateVariants(uint(productID), sess.TenantID, inputs); err != nil {
		if isDuplicateError(err) {
			renderVariantForm("SKU já existe, verifique os valores dos atributos")
		} else {
			renderVariantForm("Erro ao cadastrar variações")
		}
		return
	}

	if len(inputs) == 1 {
		ShowToast(w, "Variação cadastrada", "success")
	} else {
		ShowToast(w, fmt.Sprintf("%d variações cadastradas", len(inputs)), "success")
	}
	variants, _ := p.productStore.FindVariantsByProduct(
		uint(productID),
		sess.TenantID,
	)
	Render(templates.VariantsSection(variants, chi.URLParam(r, "id")), r, w)
}
