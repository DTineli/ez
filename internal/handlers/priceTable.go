package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
	"github.com/DTineli/ez/internal/templates/components"
	"github.com/go-chi/chi/v5"
)

func (p *ProductHandler) RenderMultiSelectTables(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tables, err := p.priceTableStore.FindAllActiveByTenant(sess.TenantID)
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
	// limpa o elemento vazio
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

	tables, err := p.priceTableStore.FindAllByTenant(sess.TenantID)
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

	fmt.Println("VAI CORINTHIANS")
	fmt.Println(errorMap)

	if len(errorMap) > 0 {
		w.Header().Set("HX-Retarget", "#price-table-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		// erro técnico → 500
		ShowToast(w, "Erro ao salvar", "error")
		Render(templates.PriceTableForm(errorMap), r, w)
		return
	}

	table := store.PriceTable{
		Name:       dto.Name,
		Percentage: dto.Percentage,
		TenantID:   sess.TenantID,
	}
	if err := p.priceTableStore.CreatePriceTable(&table); err != nil {
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
	templates.TableRow(table).Render(r.Context(), w)
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

	hasContacts, err := p.priceTableStore.HasContacts(uint(id), sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao verificar clientes", "error")
		return
	}
	if hasContacts {
		ShowToast(
			w,
			"Tabela possui clientes vinculados e não pode ser excluída",
			"error",
		)
		return
	}

	if err := p.priceTableStore.Delete(uint(id), sess.TenantID); err != nil {
		ShowToast(w, "Erro ao excluir tabela", "error")
		return
	}

	ShowToast(w, "Tabela excluída", "success")
	w.Header().Set("HX-Trigger", `{"priceTableDeleted": {}}`)
	w.WriteHeader(http.StatusOK)

}
