package handlers

import (
	"fmt"
	"net/http"

	"github.com/DTineli/ez/internal/forms"
	"github.com/DTineli/ez/internal/store"
	"github.com/DTineli/ez/internal/templates"
)

type ContactHandler struct {
	store store.ContactStore
}

func NewContactHandler(db store.ContactStore) *ContactHandler {
	return &ContactHandler{
		store: db,
	}
}

func validateContactForm(r *http.Request) (*forms.Form, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	form := forms.New(r.PostForm)
	form.Required(
		"name",
		"trade_name",
		"phone",
		"contact_type",
		"document_type",
	)

	form.MaxLength("name", 255)
	form.MaxLength("trade_name", 255)

	form.IsInt("document")
	form.IsInt("ie")
	form.IsInt("phone")
	form.IsInt("zipcode")

	form.IsEmail("email")

	return form, nil
}

func (c ContactHandler) PostNewContact(w http.ResponseWriter, r *http.Request) {
	form, err := validateContactForm(r)
	if err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	if !form.Valid() {
		ShowToast(w, "Erros de validação", "error")
		fmt.Println(form.Errors)
		_ = Render(templates.ContactForm(form, false), r, w)
		return
	}

	ShowToast(w, "Produto Cadastrado", "success")
	// form.Set("ID", strconv.Itoa(int(product.ID)))
	_ = Render(templates.ProductForm(form, true), r, w)
}

func (c ContactHandler) GetContactsPage(w http.ResponseWriter, r *http.Request) {

}

func (c ContactHandler) GetContactsForm(w http.ResponseWriter, r *http.Request) {
	Render(templates.ContactForm(forms.New(nil), false), r, w)
}
