package handlers

import (
	"net/http"

	m "github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type PriceTableHandler struct {
	priceTableStore store.priceTableStore
}

func NewPriceTableHandler(db store.PriceTableStore) *PriceTableHandler {
	return &PriceTableHandler{
		priceTableStore: db,
	}
}

func (h *PriceTableHandler) GetPriceTableForm(w http.ResponseWriter, r *http.Request) {
	err := templates.PriceTableForm().Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Erro renderizando template", http.StatusInternalServerError)
		return
	}
}

func (h *PriceTableHandler) PostNewPriceTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.FormValue("name")
	activeStr := r.FormValue("active")

	if name == "" {
		http.Error(w, "Nome é obrigatório", http.StatusBadRequest)
		return
	}

	active := activeStr == "on"

	sess := m.GetSessionFromContext(r)

	table := &store.PriceTable{
		TenantID: sess.TenantID,
		Name:     name,
		Active:   active,
	}

	err := h.priceTableStore.CreatePriceTable(table)
	if err != nil {
		http.Error(w, "Erro ao criar tabela", http.StatusInternalServerError)
		return
	}

	w.Header().Set(HXRedirect, "/tabelas-preco")
	w.WriteHeader(http.StatusOK)
}

func (h *PriceTableHandler) GetPriceTablePage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	isHX := r.Header.Get("HX-Request") == "true"

	tables, err := h.priceTableStore.FindAllByTenant(sess.TenantID)
	if err != nil {
		http.Error(w, "Erro listando tabelas", http.StatusInternalServerError)
		return
	}

	if isHX {
		err := templates.PriceTablesPage(tables).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Erro renderizando template", http.StatusInternalServerError)
		}
		return
	}

	err = templates.Layout(
		templates.PriceTablesPage(tables),
		"Ez",
		true,
		"",
	).Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Erro renderizando template", http.StatusInternalServerError)
	}
}
