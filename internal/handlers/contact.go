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
	"github.com/google/uuid"
)

type ContactHandler struct {
	contactStore    store.ContactStore
	inviteStore     store.InviteStore
	priceTableStore store.PriceTableStore
}

type NewContactHandlerParams struct {
	Contact    store.ContactStore
	Invite     store.InviteStore
	PriceTable store.PriceTableStore
}

func NewContactHandler(params NewContactHandlerParams) *ContactHandler {
	return &ContactHandler{
		contactStore:    params.Contact,
		inviteStore:     params.Invite,
		priceTableStore: params.PriceTable,
	}
}

func mapContactToForm(c *store.Contact) *forms.Form {
	form := forms.New(nil)

	form.Set("id", strconv.Itoa(int(c.ID)))
	form.Set("name", c.Name)
	form.Set("trade_name", c.TradeName)
	form.Set("contact_type", string(c.ContactType))

	form.Set("document_type", c.DocumentType)
	form.Set("document", c.Document)
	form.Set("ie", c.IE)

	form.Set("email", c.Email)
	form.Set("phone", c.Phone)

	form.Set("zipcode", c.ZipCode)
	form.Set("street", c.Street)
	form.Set("number", c.Number)
	form.Set("complement", c.Complement)
	form.Set("neighborhood", c.Neighborhood)
	form.Set("city", c.City)
	form.Set("uf", c.UF)

	form.Set("invite_link", c.InviteLink)

	form.Set("price_table_id", strconv.Itoa(int(c.PriceTableID)))

	return form
}

func validateContactForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)

	form.Set("document", strings.NewReplacer(".", "", "/", "", "-", "").Replace(form.Get("document")))
	form.Set("phone", strings.NewReplacer("(", "", ")", "", " ", "", "-", "").Replace(form.Get("phone")))

	form.Required(
		"name",
		"trade_name",
		"phone",
		"contact_type",
		"document_type",
		"price_table_id",
	)

	form.MaxLength("name", 255)
	form.MaxLength("trade_name", 255)

	form.IsInt("document")
	form.IsInt("ie")
	form.IsInt("zipcode")
	form.IsInt("price_table_id")

	form.IsEmail("email")

	return form, nil
}

func (c *ContactHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	contact, err := c.contactStore.GetOne(uint(id))
	if err != nil || contact.TenantID != sess.TenantID {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	link := &store.Invite{
		ID:       uuid.New(),
		Document: contact.Document,
		Phone:    contact.Phone,

		ContactID:    contact.ID,
		OriginTenant: sess.TenantID,
	}

	if err := c.inviteStore.Create(link); err != nil {
		ShowToast(w, "Erro ao gerar Link", "error")
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("https://%v/client/register?token=%v", r.Host, link.ID.String())

	err = c.contactStore.UpdateById(uint(contact.ID), sess.TenantID, map[string]any{
		"invite_link": url,
	})

	if err != nil {
		ShowToast(w, "Erro ao salvar contato", "error")
		return
	}

	Render(templates.InviteLink(strconv.FormatUint(id, 10), url), r, w)
}

func (c ContactHandler) fetchPriceTables(w http.ResponseWriter, tenantID uint) []store.PriceTable {
	tables, err := c.priceTableStore.FindAllByTenant(tenantID)
	if err != nil {
		return nil
	}
	return tables
}

func (c ContactHandler) PostNewContact(w http.ResponseWriter, r *http.Request) {
	form, err := validateContactForm(r)
	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	sess := m.GetSessionFromContext(r)
	priceTables := c.fetchPriceTables(w, sess.TenantID)

	if !form.Valid() {
		ShowToast(w, "Erros de validação", "error")
		fmt.Println(form.Errors)
		_ = Render(templates.ContactForm(form, false, priceTables), r, w)
		return
	}

	contact := &store.Contact{
		TenantID:    sess.TenantID,
		Name:        form.Get("name"),
		TradeName:   form.Get("trade_name"),
		ContactType: store.ContactType(form.Get("contact_type")),

		DocumentType: form.Get("document_type"),
		Document:     form.Get("document"),
		IE:           form.Get("ie"),

		Phone: form.Get("phone"),
		Email: form.Get("email"),

		ZipCode:      form.Get("zipcode"),
		Street:       form.Get("street"),
		Number:       form.Get("number"),
		Neighborhood: form.Get("neighborhood"),
		City:         form.Get("city"),
		UF:           form.Get("uf"),

		PriceTableID: uint(form.IsInt("price_table_id")),
	}

	if err := c.contactStore.CreateContact(contact); err != nil {
		ShowToast(w, "Erro ao cadastar contato", "error")
		_ = Render(templates.ContactForm(form, false, priceTables), r, w)
		return
	}

	ShowToast(w, "Contato Cadastrado", "success")
	form.Set("ID", strconv.Itoa(int(contact.ID)))
	_ = Render(templates.ContactForm(form, true, priceTables), r, w)
}

func (c ContactHandler) GetContactsForm(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)
	priceTables := c.fetchPriceTables(w, sess.TenantID)
	Render(templates.ContactForm(forms.New(nil), false, priceTables), r, w)
}

func (c ContactHandler) GetEditPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	contact, err := c.contactStore.GetOne(uint(id))
	if err != nil || contact.TenantID != sess.TenantID {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	priceTables := c.fetchPriceTables(w, sess.TenantID)
	form := mapContactToForm(contact)

	Render(templates.ContactForm(form, true, priceTables), r, w)
}

func GetPagination(r *http.Request) store.Pagination {
	page := 1
	perPage := 10
	if strPage := r.URL.Query().Get("page"); strPage != "" {
		if p, err := strconv.Atoi(strPage); err == nil && p > 0 {
			page = p
		}
	}

	return store.Pagination{Page: page, PerPage: perPage}
}

func (c ContactHandler) GetContactsPage(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	pagination := GetPagination(r)

	results, err := c.contactStore.FindAll(sess.TenantID, store.ContactFilters{
		Pagination: pagination,

		Name:        r.URL.Query().Get("name"),
		TradeName:   r.URL.Query().Get("trade_name"),
		Document:    r.URL.Query().Get("document"),
		ContactType: r.URL.Query().Get("contact_type"),
	})

	if err != nil {
		http.Error(w, "Error listing contacts", http.StatusInternalServerError)
		return
	}

	pagination.TotalPages = int(math.Floor(float64(results.Count) / float64(pagination.PerPage)))

	err = Render(templates.ContactPage(store.ListResults[store.Contact]{
		Pagination: pagination,
		Results:    *results,
	}), r, w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	form, err := validateContactForm(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	form.Set("id", strconv.Itoa(int(id)))

	priceTables := h.fetchPriceTables(w, sess.TenantID)

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		ShowToast(w, "Erro ao salvar contato", "error")
		_ = Render(templates.ContactForm(form, true, priceTables), r, w)
		return
	}

	fields := map[string]any{
		"name":           form.Get("name"),
		"trade_name":     form.Get("trade_name"),
		"contact_type":   form.Get("contact_type"),
		"document_type":  form.Get("document_type"),
		"document":       form.Get("document"),
		"ie":             form.Get("ie"),
		"email":          form.Get("email"),
		"phone":          form.Get("phone"),
		"zip_code":       form.Get("zipcode"),
		"street":         form.Get("street"),
		"number":         form.Get("number"),
		"complement":     form.Get("complement"),
		"neighborhood":   form.Get("neighborhood"),
		"city":           form.Get("city"),
		"uf":             form.Get("uf"),
		"price_table_id": form.IsInt("price_table_id"),
	}

	err = h.contactStore.UpdateById(uint(id), sess.TenantID, fields)
	if err != nil {
		ShowToast(w, "Erro ao salvar contato", "error")
		_ = Render(templates.ContactForm(form, true, priceTables), r, w)
		return
	}

	ShowToast(w, "Alterações salvas", "success")
	_ = Render(templates.ContactForm(form, true, priceTables), r, w)
}
