package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/services"
	"github.com/DTineli/ez/internal/templates"
	"github.com/DTineli/ez/internal/templates/components"
	"github.com/go-chi/chi/v5"
)

func (p *ProductHandler) RenderEditPriceTable(
	w http.ResponseWriter,
	r *http.Request) {

	sess := m.GetSessionFromContext(r)
	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	table, err := p.priceTableSvc.FindOne(uint(tableID), sess.TenantID)
	if err != nil {
		http.Error(w, "Tabela não encontrada", http.StatusNotFound)
		return
	}

	Render(templates.EditPricesPage(*table), r, w)
}

func (p *ProductHandler) SearchVariantsForPriceTable(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)
	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		Render(templates.PriceSeachTbody(nil, uint(tableID)), r, w)
		return
	}
	variants, err := p.priceTableSvc.SearchVariants(
		sess.TenantID,
		uint(tableID),
		q,
	)
	if err != nil {
		ShowToast(w, "Erro ao buscar produtos", "error")
		return
	}
	Render(templates.PriceSeachTbody(variants, uint(tableID)), r, w)
}

func (p *ProductHandler) RenderSearchPanel(
	w http.ResponseWriter,
	r *http.Request,
) {
	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	Render(templates.SearchPanel(uint(tableID)), r, w)
}

func (p *ProductHandler) CloseSearchPanel(
	w http.ResponseWriter,
	r *http.Request,
) {
	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	Render(templates.AddProductButton(uint(tableID)), r, w)
}

func (p *ProductHandler) PostProductPrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil || tableID == 0 {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil || price <= 0 {
		ShowToast(w, "Preço inválido", "error")
		return
	}

	variantID, err := strconv.ParseUint(r.FormValue("variant_id"), 10, 64)
	if err != nil || variantID == 0 {
		ShowToast(w, "Variante inválida", "error")
		return
	}

	if _, err := p.priceTableSvc.GetOne(uint(tableID), sess.TenantID); err != nil {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	priceID, err := p.priceTableSvc.AddPrice(uint(tableID), uint(variantID), price)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "Duplicate") {
			ShowToast(w, "Preço já cadastrado para essa variante nessa tabela", "error")
			return
		}
		ShowToast(w, "Erro ao salvar preço", "error")
		return
	}

	pp, err := p.priceTableSvc.GetProductPrice(priceID)
	if err != nil {
		ShowToast(w, "Preço salvo", "success")
		return
	}

	ShowToast(w, "Preço salvo", "success")
	Render(templates.PriceRow(*pp, uint(tableID)), r, w)
}

func (p *ProductHandler) PatchProductPrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil || tableID == 0 {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	priceID, err := strconv.ParseUint(chi.URLParam(r, "priceID"), 10, 64)
	if err != nil || priceID == 0 {
		ShowToast(w, "Preço inválido", "error")
		return
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil || price <= 0 {
		ShowToast(w, "Valor inválido", "error")
		return
	}

	if err := p.priceTableSvc.UpdatePrice(uint(priceID), sess.TenantID, price); err != nil {
		ShowToast(w, "Erro ao salvar preço", "error")
		return
	}

	pp, err := p.priceTableSvc.GetProductPrice(uint(priceID))
	if err != nil {
		ShowToast(w, "Preço atualizado", "success")
		return
	}

	ShowToast(w, "Preço atualizado", "success")
	Render(templates.PriceRow(*pp, uint(tableID)), r, w)
}

func (p *ProductHandler) DeleteProductPrice(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	priceID, err := strconv.ParseUint(chi.URLParam(r, "priceID"), 10, 64)
	if err != nil || priceID == 0 {
		ShowToast(w, "Preço inválido", "error")
		return
	}

	if err := p.priceTableSvc.RemovePrice(uint(priceID), sess.TenantID); err != nil {
		ShowToast(w, "Erro ao remover preço", "error")
		return
	}

	ShowToast(w, "Preço removido", "success")
	w.WriteHeader(http.StatusOK)
}

func (p *ProductHandler) RenderMultiSelectTables(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tables, err := p.priceTableSvc.FindAllActive(sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar dados", "error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	values := make([]components.MultiSelectOption, 0, len(tables))
	for _, table := range tables {
		values = append(values, components.MultiSelectOption{
			Value: strconv.Itoa(int(table.ID)),
			Label: table.Name,
		})
	}

	var selected_tables []string
	if selectedParam := r.URL.Query().Get("selected_tables"); selectedParam != "" {
		selected_tables = strings.Split(selectedParam, ",")
	}
	componentParams := components.MultiSelectParams{
		Placeholder: "Selecione uma ou mais tabelas",
		Label:       "Tabelas de Preço",
		Name:        "price_table",
		Selected:    selected_tables,
		Options:     values,
	}

	Render(components.MultiSelect(componentParams), r, w)
}

func (p *ProductHandler) GetTablePage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	tables, err := p.priceTableSvc.FindAll(sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar dados", "error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	Render(templates.PriceTablePage(tables), r, w)
}

func (p *ProductHandler) CreatePriceTable(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	dto, errorMap := formToPriceTable(r)

	if len(errorMap) > 0 {
		w.Header().Set("HX-Retarget", "#price-table-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		ShowToast(w, "Erro ao salvar", "error")
		Render(templates.PriceTableForm(errorMap), r, w)
		return
	}

	table, err := p.priceTableSvc.Create(
		sess.TenantID,
		dto.Name,
		dto.Percentage,
	)
	if err != nil {
		w.Header().Set("HX-Retarget", "#price-table-form")
		w.Header().Set("HX-Reswap", "outerHTML")

		ShowToast(w, "Erro ao salvar", "error")
		msg := err.Error()
		if strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "Duplicate") {
		}

		Render(templates.PriceTableForm(map[string]string{
			"name": "Nome ja existe",
		}), r, w)
		return
	}

	ShowToast(w, "Tabela Cadastrada", "success")
	templates.TableRow(*table).Render(r.Context(), w)
}

func (p *ProductHandler) DeletePriceTable(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	if err := p.priceTableSvc.Delete(uint(id), sess.TenantID); err != nil {
		if errors.Is(err, services.ErrPriceTableHasContacts) {
			ShowToast(
				w,
				"Tabela possui clientes vinculados e não pode ser excluída",
				"error",
			)
			return
		}
		ShowToast(w, "Erro ao excluir tabela", "error")
		return
	}

	ShowToast(w, "Tabela excluída", "success")
	w.Header().Set("HX-Trigger", `{"priceTableDeleted": {}}`)
	w.WriteHeader(http.StatusOK)
}
