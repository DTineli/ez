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

func (p *PriceTableHandler) RenderPaymentMethods(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	methods, err := p.paymentMethodSvc.FindAll(sess.TenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar dados", "error")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	Render(templates.PaymentMethodsPage(methods), r, w)
}

func (p *PriceTableHandler) CreatePaymentMethod(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		w.Header().Set("HX-Retarget", "#payment-method-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		ShowToast(w, "Erro ao salvar", "error")
		Render(templates.PaymentMethodForm(map[string]string{
			"name": "Nome é obrigatório",
		}), r, w)
		return
	}

	pm, err := p.paymentMethodSvc.Create(sess.TenantID, name)
	if err != nil {
		w.Header().Set("HX-Retarget", "#payment-method-form")
		w.Header().Set("HX-Reswap", "outerHTML")
		ShowToast(w, "Erro ao salvar", "error")

		msg := err.Error()
		errorMap := map[string]string{"name": "Erro ao salvar"}
		if strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "Duplicate") {
			errorMap["name"] = "Nome ja existe"
		}

		Render(templates.PaymentMethodForm(errorMap), r, w)
		return
	}

	ShowToast(w, "Forma de pagamento cadastrada", "success")
	Render(templates.PaymentMethodGroup(*pm), r, w)
}

func (p *PriceTableHandler) RenderPaymentTerms(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	methodID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	pm, err := p.paymentMethodSvc.GetOne(uint(methodID), sess.TenantID)
	if err != nil {
		ShowToast(w, "Forma de pagamento não encontrada", "error")
		return
	}

	p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
}

func (p *PriceTableHandler) ClosePaymentTerms(
	w http.ResponseWriter,
	r *http.Request,
) {
	methodID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	Render(templates.PaymentTermsRowEmpty(uint(methodID)), r, w)
}

func (p *PriceTableHandler) CreatePaymentTerm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	methodID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	pm, err := p.paymentMethodSvc.GetOne(uint(methodID), sess.TenantID)
	if err != nil {
		ShowToast(w, "Forma de pagamento não encontrada", "error")
		return
	}

	dueDays, errDue := strconv.Atoi(r.FormValue("due_days"))
	if errDue != nil || dueDays < 0 {
		ShowToast(w, "Dados da parcela inválidos", "error")
		p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
		return
	}

	if _, err := p.paymentMethodSvc.CreateTerm(
		sess.TenantID,
		uint(methodID),
		dueDays,
	); err != nil {
		ShowToast(w, "Erro ao salvar parcela", "error")
		p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
		return
	}

	ShowToast(w, "Parcela cadastrada", "success")
	p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
}

func (p *PriceTableHandler) DeletePaymentTerm(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	methodID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	termID, err := strconv.ParseUint(chi.URLParam(r, "termID"), 10, 64)
	if err != nil || termID == 0 {
		ShowToast(w, "Parcela inválida", "error")
		return
	}

	pm, err := p.paymentMethodSvc.GetOne(uint(methodID), sess.TenantID)
	if err != nil {
		ShowToast(w, "Forma de pagamento não encontrada", "error")
		return
	}

	if err := p.paymentMethodSvc.DeleteTerm(uint(termID), sess.TenantID); err != nil {
		ShowToast(w, "Erro ao excluir parcela", "error")
		p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
		return
	}

	ShowToast(w, "Parcela excluída", "success")
	p.renderPaymentTermsSection(w, r, *pm, sess.TenantID)
}

func (p *PriceTableHandler) renderPaymentTermsSection(
	w http.ResponseWriter,
	r *http.Request,
	pm store.PaymentMethod,
	tenantID uint,
) {
	terms, err := p.paymentMethodSvc.FindTermsByMethod(pm.ID, tenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar parcelas", "error")
		return
	}

	Render(templates.PaymentTermsRow(pm, terms), r, w)
}

func (p *PriceTableHandler) DeletePaymentMethod(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	if err := p.paymentMethodSvc.Delete(uint(id), sess.TenantID); err != nil {
		ShowToast(w, "Erro ao excluir forma de pagamento", "error")
		return
	}

	ShowToast(w, "Forma de pagamento excluída", "success")
	w.WriteHeader(http.StatusOK)
}

func (p *PriceTableHandler) RenderTablePaymentMethods(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil || tableID == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	p.renderTablePaymentMethodsSection(w, r, uint(tableID), sess.TenantID)
}

func (p *PriceTableHandler) PostTablePaymentMethod(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil || tableID == 0 {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	methodID, err := strconv.ParseUint(r.FormValue("payment_method_id"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	if err := p.priceTableSvc.AddPaymentMethod(
		uint(tableID),
		uint(methodID),
		sess.TenantID,
	); err != nil {
		ShowToast(w, "Erro ao vincular forma de pagamento", "error")
		return
	}

	ShowToast(w, "Forma de pagamento vinculada", "success")
	p.renderTablePaymentMethodsSection(w, r, uint(tableID), sess.TenantID)
}

func (p *PriceTableHandler) DeleteTablePaymentMethod(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	tableID, err := strconv.ParseUint(chi.URLParam(r, "tableID"), 10, 64)
	if err != nil || tableID == 0 {
		ShowToast(w, "Tabela inválida", "error")
		return
	}

	methodID, err := strconv.ParseUint(chi.URLParam(r, "methodID"), 10, 64)
	if err != nil || methodID == 0 {
		ShowToast(w, "Forma de pagamento inválida", "error")
		return
	}

	if err := p.priceTableSvc.RemovePaymentMethod(
		uint(tableID),
		uint(methodID),
		sess.TenantID,
	); err != nil {
		ShowToast(w, "Erro ao remover forma de pagamento", "error")
		return
	}

	ShowToast(w, "Forma de pagamento removida", "success")
	p.renderTablePaymentMethodsSection(w, r, uint(tableID), sess.TenantID)
}

func (p *PriceTableHandler) renderTablePaymentMethodsSection(
	w http.ResponseWriter,
	r *http.Request,
	tableID, tenantID uint,
) {
	linked, err := p.paymentMethodSvc.FindAllByPriceTable(tableID, tenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar formas de pagamento", "error")
		return
	}

	all, err := p.paymentMethodSvc.FindAll(tenantID)
	if err != nil {
		ShowToast(w, "Erro ao recuperar formas de pagamento", "error")
		return
	}

	linkedIDs := make(map[uint]bool, len(linked))
	for _, pm := range linked {
		linkedIDs[pm.ID] = true
	}
	available := make([]store.PaymentMethod, 0, len(all))
	for _, pm := range all {
		if !linkedIDs[pm.ID] {
			available = append(available, pm)
		}
	}

	Render(
		templates.TablePaymentMethodsSection(tableID, linked, available),
		r,
		w,
	)
}
