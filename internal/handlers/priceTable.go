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

	table, err := p.priceTableSvc.Create(sess.TenantID, dto.Name, dto.Percentage)
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
