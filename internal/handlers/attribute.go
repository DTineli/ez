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
