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

	url := fmt.Sprintf(
		"https://%v/client/register?token=%v",
		r.Host,
		link.ID.String(),
	)

	err = c.contactStore.UpdateById(
		uint(contact.ID),
		sess.TenantID,
		map[string]any{
			"invite_link": url,
		},
	)

	if err != nil {
		ShowToast(w, "Erro ao salvar contato", "error")
		return
	}

	Render(templates.InviteLink(strconv.FormatUint(id, 10), url), r, w)
}

func (c ContactHandler) PostNewContact(w http.ResponseWriter, r *http.Request) {
	form, err := validateContactForm(r)
	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	sess := m.GetSessionFromContext(r)
	if !form.Valid() {
		_ = Render(templates.ContactForm(form, false), r, w)
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

		PriceTables: convertPriceTable(r.Form["price_table"]),
	}

	if err := c.contactStore.CreateContact(contact); err != nil {
		form.Errors.Add(
			"general",
			"Erro ao cadastrar contato. Tente novamente.",
		)
		_ = Render(templates.ContactForm(form, false), r, w)
		return
	}

	ShowToast(w, "Contato Cadastrado", "success")
	form.Set("ID", strconv.Itoa(int(contact.ID)))

	form.Set("price_table", joinPriceTableIDs(contact.PriceTables))

	_ = Render(templates.ContactForm(form, true), r, w)
}

func (c ContactHandler) GetContactsForm(
	w http.ResponseWriter,
	r *http.Request,
) {
	Render(templates.ContactForm(forms.New(nil), false), r, w)
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

	form := mapContactToForm(contact)
	Render(templates.ContactForm(form, true), r, w)
}

func (c ContactHandler) GetContactsPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	sess := m.GetSessionFromContext(r)

	pagination := GetPagination(r)

	results, err := c.contactStore.FindAll(sess.TenantID, store.ContactFilters{
		Pagination: pagination,

		Name:        strings.TrimSpace(r.URL.Query().Get("name")),
		TradeName:   strings.TrimSpace(r.URL.Query().Get("trade_name")),
		Document:    strings.TrimSpace(r.URL.Query().Get("document")),
		ContactType: strings.TrimSpace(r.URL.Query().Get("contact_type")),
	})

	if err != nil {
		http.Error(w, "Error listing contacts", http.StatusInternalServerError)
		return
	}

	pagination.TotalPages = int(
		math.Floor(float64(results.Count) / float64(pagination.PerPage)),
	)

	err = Render(templates.ContactPage(store.ListResults[store.Contact]{
		Pagination: pagination,
		Results:    *results,
	}), r, w)

	if err != nil {
		http.Error(
			w,
			"Error rendering template",
			http.StatusInternalServerError,
		)
	}
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	sess := m.GetSessionFromContext(r)

	form, err := validateContactForm(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	form.Set("id", strconv.Itoa(int(id)))

	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		_ = Render(templates.ContactForm(form, true), r, w)
		return
	}

	tableJoin := strings.Join(r.Form["price_table"], ",")
	tableIds := parsePriceTableIDs(tableJoin)

	fields := map[string]any{
		"name":          form.Get("name"),
		"trade_name":    form.Get("trade_name"),
		"contact_type":  form.Get("contact_type"),
		"document_type": form.Get("document_type"),
		"document":      form.Get("document"),
		"ie":            form.Get("ie"),
		"email":         form.Get("email"),
		"phone":         form.Get("phone"),
		"zip_code":      form.Get("zipcode"),
		"street":        form.Get("street"),
		"number":        form.Get("number"),
		"complement":    form.Get("complement"),
		"neighborhood":  form.Get("neighborhood"),
		"city":          form.Get("city"),
		"uf":            form.Get("uf"),

		"price_table_ids": tableIds,
	}

	err = h.contactStore.UpdateById(uint(id), sess.TenantID, fields)

	form.Set("price_table", tableJoin)
	if err != nil {
		form.Errors.Add("general", "Erro ao salvar contato. Tente novamente.")
		_ = Render(templates.ContactForm(form, true), r, w)
		return
	}

	ShowToast(w, "Alterações salvas", "success")

	_ = Render(templates.ContactForm(form, true), r, w)
}
